package finance

import (
	"context"
	"errors"
	"fmt"

	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo *Repository
	db   *db.DB
}

func NewService(repo *Repository, d *db.DB) *Service {
	return &Service{repo: repo, db: d}
}

func (s *Service) HandleOrderPlaced(ctx context.Context, e events.Event) error {
	payload, ok := e.Payload.(events.OrderPlacedPayload)
	if !ok {
		return fmt.Errorf("finance: unexpected payload type for %s", events.TypeOrderPlaced)
	}

	var total float64
	for _, item := range payload.Items {
		total += float64(item.Quantity) * item.UnitPrice
	}

	err := s.db.WithTenant(ctx, e.TenantID, func(tx pgx.Tx) error {
		inv, err := s.repo.CreateInvoice(ctx, tx, e.TenantID, payload.OrderID, total)
		if err != nil {
			return fmt.Errorf("repo create invoice record: %w", err)
		}
		if _, err := s.repo.CreateLedgerEntry(ctx, tx, e.TenantID, inv.InvoiceID, "debit", total); err != nil {
			return fmt.Errorf("repo ledger entry debit tracking execution: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("handling order placed event pipeline: %w", err)
	}
	return nil
}

func (s *Service) ListInvoices(ctx context.Context, tenantID uuid.UUID) ([]InvoiceResponse, error) {
	var out []InvoiceResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		invoices, err := s.repo.ListInvoices(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo fetch invoice list records: %w", err)
		}
		out = make([]InvoiceResponse, 0, len(invoices))
		for _, inv := range invoices {
			out = append(out, toInvoiceResponse(inv))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing invoices: %w", err)
	}
	return out, nil
}

func (s *Service) GetInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) (*InvoiceResponse, error) {
	var resp InvoiceResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		inv, err := s.repo.GetInvoice(ctx, tx, tenantID, invoiceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvoiceNotFound
			}
			return fmt.Errorf("repo fetch unique invoice profile: %w", err)
		}
		resp = toInvoiceResponse(*inv)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// PayInvoice marks an invoice paid and records the offsetting credit ledger entry
func (s *Service) PayInvoice(ctx context.Context, tenantID, invoiceID uuid.UUID) (*InvoiceResponse, error) {
	var resp InvoiceResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		inv, err := s.repo.GetInvoice(ctx, tx, tenantID, invoiceID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvoiceNotFound
			}
			return fmt.Errorf("lookup settlement targeting ledger: %w", err)
		}

		if err := s.repo.UpdateInvoiceStatus(ctx, tx, tenantID, invoiceID, "paid"); err != nil {
			return fmt.Errorf("repo modification balance shift execution: %w", err)
		}
		if _, err := s.repo.CreateLedgerEntry(ctx, tx, tenantID, invoiceID, "credit", inv.Amount); err != nil {
			return fmt.Errorf("repo ledger entry settlement recording execution: %w", err)
		}

		inv.Status = "paid"
		resp = toInvoiceResponse(*inv)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func toInvoiceResponse(inv Invoice) InvoiceResponse {
	return InvoiceResponse{
		InvoiceID: inv.InvoiceID.String(),
		OrderID:   inv.OrderID.String(),
		Amount:    inv.Amount,
		Status:    inv.Status,
		IssuedAt:  inv.IssuedAt,
	}
}
