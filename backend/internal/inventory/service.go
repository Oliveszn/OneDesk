package inventory

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var ErrInvalidAdjustment = errors.New("stock adjustment would result in a negative quantity")

// stockLowCandidate is an internal bookkeeping type — a product/warehouse
// pair whose quantity crossed below its reorder point during an
// adjustment, collected during a transaction and published as stock.low
// events AFTER that transaction commits (see the publish calls below for
// why: never publish from inside a still-open transaction whose outcome
// isn't final yet).
type stockLowCandidate struct {
	ProductID    uuid.UUID
	WarehouseID  uuid.UUID
	ReorderPoint int
}

type Service struct {
	repo    *Repository
	billing *billing.Service
	bus     *events.Bus
	db      *db.DB
}

func NewService(repo *Repository, b *billing.Service, bus *events.Bus, d *db.DB) *Service {
	return &Service{repo: repo, billing: b, bus: bus, db: d}
}

func (s *Service) CreateWarehouse(ctx context.Context, tenantID uuid.UUID, name string) (*WarehouseResponse, error) {
	if name == "" {
		return nil, errors.New("name is required")
	}

	var resp WarehouseResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		w, err := s.repo.CreateWarehouse(ctx, tx, tenantID, name)
		if err != nil {
			return fmt.Errorf("repo create warehouse execution: %w", err)
		}
		resp = WarehouseResponse{WarehouseID: w.WarehouseID.String(), Name: w.Name}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating warehouse: %w", err)
	}
	return &resp, nil
}

func (s *Service) ListWarehouses(ctx context.Context, tenantID uuid.UUID) ([]WarehouseResponse, error) {
	var out []WarehouseResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		warehouses, err := s.repo.ListWarehouses(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo list warehouses execution: %w", err)
		}
		out = make([]WarehouseResponse, 0, len(warehouses))
		for _, w := range warehouses {
			out = append(out, WarehouseResponse{WarehouseID: w.WarehouseID.String(), Name: w.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing warehouses: %w", err)
	}
	return out, nil
}

func (s *Service) CreateProduct(ctx context.Context, tenantID uuid.UUID, req CreateProductRequest) (*ProductResponse, error) {
	if req.SKU == "" || req.Name == "" {
		return nil, errors.New("sku and name are required")
	}

	var resp ProductResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		if err := s.billing.ConsumeEntitlement(ctx, tx, tenantID, billing.ResourceProducts); err != nil {
			return fmt.Errorf("entitlement check boundary: %w", err)
		}

		p, err := s.repo.CreateProduct(ctx, tx, tenantID, req.SKU, req.Name)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isDuplicateKeyError(err) {
				return ErrSKUTaken
			}
			return fmt.Errorf("repo insert product execution: %w", err)
		}
		resp = ProductResponse{ProductID: p.ProductID.String(), SKU: p.SKU, Name: p.Name}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Service) ListProducts(ctx context.Context, tenantID uuid.UUID) ([]ProductResponse, error) {
	var out []ProductResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		products, err := s.repo.ListProducts(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo list products execution: %w", err)
		}
		out = make([]ProductResponse, 0, len(products))
		for _, p := range products {
			out = append(out, ProductResponse{ProductID: p.ProductID.String(), SKU: p.SKU, Name: p.Name})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	return out, nil
}

func (s *Service) AdjustStock(ctx context.Context, tenantID, productID uuid.UUID, req AdjustStockRequest) (*StockLevelResponse, error) {
	warehouseID, err := uuid.Parse(req.WarehouseID)
	if err != nil {
		return nil, errors.New("invalid warehouse_id")
	}

	var resp StockLevelResponse
	var lowStock *stockLowCandidate

	err = s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {

		if _, err := s.repo.GetProduct(ctx, tx, tenantID, productID); err != nil {
			return err
		}

		newQty, reorderPoint, ok, err := s.repo.AdjustStock(ctx, tx, tenantID, productID, warehouseID, req.Delta)
		if err != nil {
			return err
		}
		if !ok {
			return ErrInvalidAdjustment
		}
		if reorderPoint > 0 && newQty < reorderPoint {
			lowStock = &stockLowCandidate{ProductID: productID, WarehouseID: warehouseID, ReorderPoint: reorderPoint}
		}

		resp = StockLevelResponse{
			ProductID:   productID.String(),
			WarehouseID: warehouseID.String(),
			Quantity:    newQty,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if lowStock != nil {
		s.publishStockLow(ctx, tenantID, *lowStock)
	}

	return &resp, nil
}

func (s *Service) GetStockLevels(ctx context.Context, tenantID, productID uuid.UUID) ([]StockLevelResponse, error) {
	var out []StockLevelResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		if _, err := s.repo.GetProduct(ctx, tx, tenantID, productID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrProductNotFound
			}
			return fmt.Errorf("lookup verification baseline: %w", err)
		}

		levels, err := s.repo.GetStockLevels(ctx, tx, tenantID, productID)
		if err != nil {
			return fmt.Errorf("repo collection retrieval execution: %w", err)
		}
		out = make([]StockLevelResponse, 0, len(levels))
		for _, l := range levels {
			out = append(out, StockLevelResponse{
				ProductID:    l.ProductID.String(),
				WarehouseID:  l.WarehouseID.String(),
				Quantity:     l.Quantity,
				ReorderPoint: l.ReorderPoint,
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) HandleOrderPlaced(ctx context.Context, e events.Event) error {
	payload, ok := e.Payload.(events.OrderPlacedPayload)
	if !ok {
		return fmt.Errorf("inventory: unexpected payload type for %s", events.TypeOrderPlaced)
	}

	var lowStock []stockLowCandidate

	err := s.db.WithTenant(ctx, e.TenantID, func(tx pgx.Tx) error {
		for _, item := range payload.Items {
			if _, err := s.repo.GetProduct(ctx, tx, e.TenantID, item.ProductID); err != nil {
				return fmt.Errorf("product %s: %w", item.ProductID, err)
			}

			newQty, reorderPoint, ok, err := s.repo.AdjustStock(ctx, tx, e.TenantID, item.ProductID, item.WarehouseID, -item.Quantity)
			if err != nil {
				return fmt.Errorf("product %s: %w", item.ProductID, err)
			}
			if !ok {
				return fmt.Errorf("product %s: %w", item.ProductID, events.ErrInsufficientStock)
			}
			if reorderPoint > 0 && newQty < reorderPoint {
				lowStock = append(lowStock, stockLowCandidate{ProductID: item.ProductID, WarehouseID: item.WarehouseID, ReorderPoint: reorderPoint})
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, c := range lowStock {
		s.publishStockLow(ctx, e.TenantID, c)
	}
	return nil
}

// HandlePOReceived is registered as a subscriber to events.TypePOReceived
// in cmd/api/main.go — a purchase order being marked received restocks
// each of its line items. Positive deltas can't fail the "would go
// negative" check, so an !ok result here would mean something is
// actually wrong (e.g. a genuine race with a concurrent decrement
// draining the row simultaneously) rather than an expected outcome —
// worth surfacing as a real error rather than silently swallowing it.
func (s *Service) HandlePOReceived(ctx context.Context, e events.Event) error {
	payload, ok := e.Payload.(events.POReceivedPayload)
	if !ok {
		return fmt.Errorf("inventory: unexpected payload type for %s", events.TypePOReceived)
	}

	return s.db.WithTenant(ctx, e.TenantID, func(tx pgx.Tx) error {
		for _, item := range payload.Items {
			if _, err := s.repo.GetProduct(ctx, tx, e.TenantID, item.ProductID); err != nil {
				return fmt.Errorf("product %s: %w", item.ProductID, err)
			}

			_, _, ok, err := s.repo.AdjustStock(ctx, tx, e.TenantID, item.ProductID, item.WarehouseID, item.Quantity)
			if err != nil {
				return fmt.Errorf("product %s: %w", item.ProductID, err)
			}
			if !ok {
				return fmt.Errorf("product %s: restock failed unexpectedly (positive delta should never be rejected)", item.ProductID)
			}
		}
		return nil
	})
}

// publishStockLow is deliberately best-effort: the stock adjustment it's
// reporting on has ALREADY committed successfully by the time this runs.
// A failure to notify Procurement doesn't mean the adjustment failed —
// it means a suggestion didn't get raised, which is a worse outcome for
// the tenant than "the adjustment silently failed" would be, but not
// one that should make an otherwise-successful stock adjustment or order
// look like it failed. Logged, not returned, on purpose.
func (s *Service) publishStockLow(ctx context.Context, tenantID uuid.UUID, c stockLowCandidate) {
	err := s.bus.Publish(ctx, events.Event{
		Type:     events.TypeStockLow,
		TenantID: tenantID,
		Payload: events.StockLowPayload{
			ProductID:         c.ProductID,
			WarehouseID:       c.WarehouseID,
			SuggestedQuantity: c.ReorderPoint,
		},
	})
	if err != nil {
		log.Printf("inventory: publishing stock.low for product %s failed: %v", c.ProductID, err)
	}
}

func isDuplicateKeyError(err error) bool {
	return false
}
