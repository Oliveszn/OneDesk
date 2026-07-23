package sales

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/cache"
	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const cacheTTL = 5 * time.Minute

var ErrEmptyOrder = errors.New("an order must have at least one item")

// Service depends on billing.service to consume teneats order and events.bus to publish order.placed
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

func (s *Service) CreateCustomer(ctx context.Context, tenantID uuid.UUID, req CreateCustomerRequest) (*CustomerResponse, error) {
	if req.Name == "" {
		return nil, errors.New("name is required")
	}

	var resp CustomerResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		c, err := s.repo.CreateCustomer(ctx, tx, tenantID, req.Name, req.Email)
		if err != nil {
			return fmt.Errorf("repo save customer: %w", err)
		}
		resp = CustomerResponse{CustomerID: c.CustomerID.String(), Name: c.Name, Email: c.Email}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("creating customer: %w", err)
	}
	s.cache.Delete(ctx, cache.TenantKey(tenantID, "customers", "list"))
	return &resp, nil
}

func (s *Service) ListCustomers(ctx context.Context, tenantID uuid.UUID) ([]CustomerResponse, error) {
	key := cache.TenantKey(tenantID, "customers", "list")

	var cached []CustomerResponse
	if hit, _ := s.cache.Get(ctx, key, &cached); hit {
		return cached, nil
	}

	var out []CustomerResponse
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		customers, err := s.repo.ListCustomers(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("repo list customers: %w", err)
		}
		out = make([]CustomerResponse, 0, len(customers))
		for _, c := range customers {
			out = append(out, CustomerResponse{CustomerID: c.CustomerID.String(), Name: c.Name, Email: c.Email})
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing customers: %w", err)
	}
	s.cache.Set(ctx, key, out, cacheTTL)
	return out, nil
}

func (s *Service) PlaceOrder(ctx context.Context, tenantID uuid.UUID, req CreateOrderRequest) (*OrderResponse, error) {
	customerID, err := uuid.Parse(req.CustomerID)
	if err != nil {
		return nil, errors.New("invalid customer_id")
	}
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	type parsedItem struct {
		ProductID   uuid.UUID
		WarehouseID uuid.UUID
		Quantity    int
		UnitPrice   float64
	}
	parsedItems := make([]parsedItem, len(req.Items))
	for i, it := range req.Items {
		productID, err := uuid.Parse(it.ProductID)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid product_id", i)
		}
		warehouseID, err := uuid.Parse(it.WarehouseID)
		if err != nil {
			return nil, fmt.Errorf("item %d: invalid warehouse_id", i)
		}
		if it.Quantity <= 0 {
			return nil, fmt.Errorf("item %d: quantity must be positive", i)
		}
		parsedItems[i] = parsedItem{productID, warehouseID, it.Quantity, it.UnitPrice}
	}

	var order *Order
	var items []OrderItem

	err = s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		if err := s.billing.ConsumeEntitlement(ctx, tx, tenantID, billing.ResourceOrders); err != nil {
			return fmt.Errorf("entitlement validation boundary: %w", err)
		}

		o, err := s.repo.CreateOrder(ctx, tx, tenantID, customerID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) || isForeignKeyViolation(err, "customer") {
				return ErrCustomerNotFound
			}
			return fmt.Errorf("repo insert order execution: %w", err)
		}
		order = o

		for _, it := range parsedItems {
			item, err := s.repo.CreateOrderItem(ctx, tx, tenantID, order.OrderID, it.ProductID, it.WarehouseID, it.Quantity, it.UnitPrice)
			if err != nil {
				if isForeignKeyViolation(err, "product") {
					return ErrInvalidProduct
				}
				if isForeignKeyViolation(err, "warehouse") {
					return ErrInvalidWarehouse
				}
				return fmt.Errorf("repo line-item persistence execution: %w", err)
			}
			items = append(items, *item)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Order is committed. Publish fulfillment as a separate step
	eventItems := make([]events.OrderPlacedItem, len(items))
	for i, it := range items {
		eventItems[i] = events.OrderPlacedItem{
			ProductID:   it.ProductID,
			WarehouseID: it.WarehouseID,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		}
	}

	publishErr := s.bus.Publish(ctx, events.Event{
		Type:     events.TypeOrderPlaced,
		TenantID: tenantID,
		Payload:  events.OrderPlacedPayload{OrderID: order.OrderID, Items: eventItems},
	})
	if publishErr != nil {
		if compErr := s.markStockIssue(ctx, tenantID, order.OrderID); compErr != nil {
			return nil, fmt.Errorf("fulfillment failed (%w) AND marking order as stock_issue failed (%w)", publishErr, compErr)
		}
		return nil, fmt.Errorf("order %s created but fulfillment failed, marked stock_issue: %w", order.OrderID, publishErr)
	}

	return buildOrderResponse(order, items), nil
}

func (s *Service) markStockIssue(ctx context.Context, tenantID, orderID uuid.UUID) error {
	return s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		return s.repo.UpdateOrderStatus(ctx, tx, tenantID, orderID, "stock_issue")
	})
}

func (s *Service) GetOrder(ctx context.Context, tenantID, orderID uuid.UUID) (*OrderResponse, error) {
	var order *Order
	var items []OrderItem

	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		o, err := s.repo.GetOrder(ctx, tx, tenantID, orderID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrOrderNotFound
			}
			return fmt.Errorf("repo get order execution: %w", err)
		}
		order = o

		its, err := s.repo.ListOrderItems(ctx, tx, tenantID, orderID)
		if err != nil {
			return fmt.Errorf("repo list items execution: %w", err)
		}
		items = its
		return nil
	})
	if err != nil {
		return nil, err
	}

	return buildOrderResponse(order, items), nil
}

func buildOrderResponse(o *Order, items []OrderItem) *OrderResponse {
	resp := &OrderResponse{
		OrderID:    o.OrderID.String(),
		CustomerID: o.CustomerID.String(),
		Status:     o.Status,
		Items:      make([]OrderItemResponse, 0, len(items)),
	}
	var total float64
	for _, it := range items {
		resp.Items = append(resp.Items, OrderItemResponse{
			ProductID:   it.ProductID.String(),
			WarehouseID: it.WarehouseID.String(),
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
		total += float64(it.Quantity) * it.UnitPrice
	}
	resp.Total = total
	return resp
}

func isForeignKeyViolation(err error, targetKeyword string) bool {
	// Custom database driver error code check could be wired here if needed
	return false
}
