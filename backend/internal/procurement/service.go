package procurement

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
	bus  *events.Bus
	db   *db.DB
}

func NewService(repo *Repository, bus *events.Bus, d *db.DB) *Service {
	return &Service{repo: repo, bus: bus, db: d}
}

func (s *Service) CreateVendor(ctx context.Context, tenantID uuid.UUID, name string) (*VendorResponse, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}
	var resp VendorResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		v, err := s.repo.CreateVendor(ctx, tx, tenantID, name)
		if err != nil {
			return fmt.Errorf("repo save vendor metrics: %w", err)
		}
		resp = VendorResponse{VendorID: v.VendorID.String(), Name: v.Name}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating vendor: %w", err)
	}
	return &resp, nil
}

func (s *Service) ListVendors(ctx context.Context, tenantID uuid.UUID) ([]VendorResponse, error) {
	var out []VendorResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		vendors, err := s.repo.ListVendors(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo list workspace vendors: %w", err)
		}
		out = make([]VendorResponse, 0, len(vendors))
		for _, v := range vendors {
			out = append(out, VendorResponse{VendorID: v.VendorID.String(), Name: v.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing vendors: %w", err)
	}
	return out, nil
}

func (s *Service) HandleStockLow(ctx context.Context, e events.Event) error {
	payload, ok := e.Payload.(events.StockLowPayload)
	if !ok {
		return fmt.Errorf("procurement: unexpected payload type for %s", events.TypeStockLow)
	}
	if payload.SuggestedQuantity <= 0 {
		return nil // nothing sensible to order
	}

	err := s.db.WithTenant(ctx, e.TenantID, func(tx pgx.Tx) error {
		po, err := s.repo.CreatePurchaseOrder(ctx, tx, e.TenantID)
		if err != nil {
			return fmt.Errorf("repo initialize auto purchase order: %w", err)
		}
		_, err = s.repo.CreatePOItem(ctx, tx, e.TenantID, po.POID, payload.ProductID, payload.WarehouseID, payload.SuggestedQuantity)
		if err != nil {
			return fmt.Errorf("repo assign item to auto purchase order: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("handling stock low event pipeline: %w", err)
	}
	return nil
}

func (s *Service) ListPurchaseOrders(ctx context.Context, tenantID uuid.UUID) ([]PurchaseOrderResponse, error) {
	var out []PurchaseOrderResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		pos, err := s.repo.ListPurchaseOrders(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo query purchase orders map: %w", err)
		}
		out = make([]PurchaseOrderResponse, 0, len(pos))
		for _, po := range pos {
			items, err := s.repo.ListPOItems(ctx, tx, tenantID, po.POID)
			if err != nil {
				return fmt.Errorf("repo list purchase order line items: %w", err)
			}
			out = append(out, toPurchaseOrderResponse(po, items))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing purchase orders: %w", err)
	}
	return out, nil
}

func (s *Service) GetPurchaseOrder(ctx context.Context, tenantID, poID uuid.UUID) (*PurchaseOrderResponse, error) {
	var resp PurchaseOrderResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		po, err := s.repo.GetPurchaseOrder(ctx, tx, tenantID, poID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrPONotFound
			}
			return fmt.Errorf("repo extract unique purchase order: %w", err)
		}
		items, err := s.repo.ListPOItems(ctx, tx, tenantID, poID)
		if err != nil {
			return fmt.Errorf("repo line-item mapping resolution: %w", err)
		}
		resp = toPurchaseOrderResponse(*po, items)
		return nil
	})
	if err != nil {
		return nil, err // Bubbled straight to let handler analyze sentinels directly
	}
	return &resp, nil
}

func (s *Service) AssignVendorAndSend(ctx context.Context, tenantID, poID, vendorID uuid.UUID) (*PurchaseOrderResponse, error) {
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		err := s.repo.AssignVendorAndSend(ctx, tx, tenantID, poID, vendorID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidTransition
			}
			return fmt.Errorf("repo execution on purchase order state update: %w", err)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidTransition) {
			if _, getErr := s.GetPurchaseOrder(ctx, tenantID, poID); errors.Is(getErr, ErrPONotFound) {
				return nil, ErrPONotFound
			}
			return nil, ErrInvalidTransition
		}
		return nil, err // ErrVendorNotFound or unexpected context blocks
	}
	return s.GetPurchaseOrder(ctx, tenantID, poID)
}

func (s *Service) ReceivePurchaseOrder(ctx context.Context, tenantID, poID uuid.UUID) (*PurchaseOrderResponse, error) {
	var items []POItem

	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		if err := s.repo.MarkReceived(ctx, tx, tenantID, poID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrInvalidTransition // Not in 'sent' state or missing completely
			}
			return fmt.Errorf("repo update to received status: %w", err)
		}
		its, err := s.repo.ListPOItems(ctx, tx, tenantID, poID)
		if err != nil {
			return fmt.Errorf("repo load delivery tracking components: %w", err)
		}
		items = its
		return nil
	})
	if err != nil {
		// Differentiate missing completely from bad current transition path state
		if errors.Is(err, ErrInvalidTransition) {
			if _, getErr := s.GetPurchaseOrder(ctx, tenantID, poID); errors.Is(getErr, ErrPONotFound) {
				return nil, ErrPONotFound
			}
		}
		return nil, err
	}

	eventItems := make([]events.POReceivedItem, len(items))
	for i, it := range items {
		eventItems[i] = events.POReceivedItem{
			ProductID:   it.ProductID,
			WarehouseID: it.WarehouseID,
			Quantity:    it.Quantity,
		}
	}

	publishErr := s.bus.Publish(ctx, events.Event{
		Type:     events.TypePOReceived,
		TenantID: tenantID,
		Payload:  events.POReceivedPayload{POID: poID, Items: eventItems},
	})
	if publishErr != nil {
		if compErr := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
			return s.repo.MarkReceiveIssue(ctx, tx, tenantID, poID)
		}); compErr != nil {
			return nil, fmt.Errorf("restock failed (%w) AND marking receive_issue failed (%w)", publishErr, compErr)
		}
		return nil, fmt.Errorf("purchase order %s marked received but restock failed, marked receive_issue: %w", poID, publishErr)
	}

	return s.GetPurchaseOrder(ctx, tenantID, poID)
}

func toPurchaseOrderResponse(po PurchaseOrder, items []POItem) PurchaseOrderResponse {
	resp := PurchaseOrderResponse{
		POID:   po.POID.String(),
		Status: po.Status,
		Items:  make([]POItemResponse, 0, len(items)),
	}
	if po.HasVendor() {
		vid := po.VendorID.String()
		resp.VendorID = &vid
	}
	for _, it := range items {
		resp.Items = append(resp.Items, POItemResponse{
			ProductID:   it.ProductID.String(),
			WarehouseID: it.WarehouseID.String(),
			Quantity:    it.Quantity,
		})
	}
	return resp
}
