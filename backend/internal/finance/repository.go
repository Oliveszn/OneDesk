package finance

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInvoiceNotFound = errors.New("invoice not found")

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateInvoice(ctx context.Context, tx pgx.Tx, tenantID, orderID uuid.UUID, amount float64) (*Invoice, error) {
	var inv Invoice
	err := tx.QueryRow(ctx,
		`INSERT INTO invoices (tenant_id, order_id, amount, status)
		 VALUES ($1, $2, $3, 'unpaid')
		 RETURNING invoice_id, tenant_id, order_id, amount, status, issued_at`,
		tenantID, orderID, amount,
	).Scan(&inv.InvoiceID, &inv.TenantID, &inv.OrderID, &inv.Amount, &inv.Status, &inv.IssuedAt)
	if err != nil {
		return nil, fmt.Errorf("creating invoice: %w", err)
	}
	return &inv, nil
}

func (r *Repository) CreateLedgerEntry(ctx context.Context, tx pgx.Tx, tenantID, invoiceID uuid.UUID, entryType string, amount float64) (*LedgerEntry, error) {
	var e LedgerEntry
	err := tx.QueryRow(ctx,
		`INSERT INTO ledger_entries (tenant_id, invoice_id, entry_type, amount)
		 VALUES ($1, $2, $3, $4)
		 RETURNING entry_id, tenant_id, invoice_id, entry_type, amount, created_at`,
		tenantID, invoiceID, entryType, amount,
	).Scan(&e.EntryID, &e.TenantID, &e.InvoiceID, &e.EntryType, &e.Amount, &e.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating ledger entry: %w", err)
	}
	return &e, nil
}

func (r *Repository) GetInvoice(ctx context.Context, tx pgx.Tx, tenantID, invoiceID uuid.UUID) (*Invoice, error) {
	var inv Invoice
	err := tx.QueryRow(ctx,
		`SELECT invoice_id, tenant_id, order_id, amount, status, issued_at
		 FROM invoices WHERE tenant_id = $1 AND invoice_id = $2`,
		tenantID, invoiceID,
	).Scan(&inv.InvoiceID, &inv.TenantID, &inv.OrderID, &inv.Amount, &inv.Status, &inv.IssuedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvoiceNotFound
		}
		return nil, fmt.Errorf("getting invoice: %w", err)
	}
	return &inv, nil
}

func (r *Repository) ListInvoices(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]Invoice, error) {
	rows, err := tx.Query(ctx,
		`SELECT invoice_id, tenant_id, order_id, amount, status, issued_at
		 FROM invoices WHERE tenant_id = $1 ORDER BY issued_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing invoices: %w", err)
	}
	defer rows.Close()

	var out []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.InvoiceID, &inv.TenantID, &inv.OrderID, &inv.Amount, &inv.Status, &inv.IssuedAt); err != nil {
			return nil, fmt.Errorf("scanning invoice: %w", err)
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}

func (r *Repository) UpdateInvoiceStatus(ctx context.Context, tx pgx.Tx, tenantID, invoiceID uuid.UUID, status string) error {
	tag, err := tx.Exec(ctx,
		`UPDATE invoices SET status = $1 WHERE tenant_id = $2 AND invoice_id = $3`,
		status, tenantID, invoiceID,
	)
	if err != nil {
		return fmt.Errorf("updating invoice status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInvoiceNotFound
	}
	return nil
}
