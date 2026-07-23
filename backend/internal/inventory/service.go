package inventory

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/cache"
	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const cacheTTL = 5 * time.Minute

var ErrInvalidAdjustment = errors.New("stock adjustment would result in a negative quantity")

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
	cache   *cache.Client
}

func NewService(repo *Repository, b *billing.Service, bus *events.Bus, d *db.DB, c *cache.Client) *Service {
	return &Service{repo: repo, billing: b, bus: bus, db: d, cache: c}
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

	s.cache.Delete(ctx, cache.TenantKey(tenantID, "warehouses", "list"))
	return &resp, nil
}

func (s *Service) ListWarehouses(ctx context.Context, tenantID uuid.UUID) ([]WarehouseResponse, error) {
	key := cache.TenantKey(tenantID, "warehouses", "list")

	var cached []WarehouseResponse
	if hit, _ := s.cache.Get(ctx, key, &cached); hit {
		return cached, nil
	}

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
	s.cache.Set(ctx, key, out, cacheTTL)
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
	s.cache.Delete(ctx, cache.TenantKey(tenantID, "products", "list"))
	return &resp, nil
}

func (s *Service) ListProducts(ctx context.Context, tenantID uuid.UUID) ([]ProductResponse, error) {
	key := cache.TenantKey(tenantID, "products", "list")

	var cached []ProductResponse
	if hit, _ := s.cache.Get(ctx, key, &cached); hit {
		return cached, nil
	}

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
	s.cache.Set(ctx, key, out, cacheTTL)
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
