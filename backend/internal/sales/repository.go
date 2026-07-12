package sales

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrCustomerNotFound = errors.New("customer not found")
	ErrInvalidProduct   = errors.New("product not found or does not belong to this tenant")
	ErrInvalidWarehouse = errors.New("warehouse not found or does not belong to this tenant")
	ErrOrderNotFound    = errors.New("order not found")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateCustomer(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, name, email string) (*Customer, error) {
	var c Customer
	err := tx.QueryRow(ctx,
		`INSERT INTO customers (tenant_id, name, email) VALUES ($1, $2, $3)
		 RETURNING customer_id, tenant_id, name, email, created_at`,
		tenantID, name, email,
	).Scan(&c.CustomerID, &c.TenantID, &c.Name, &c.Email, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating customer: %w", err)
	}
	return &c, nil
}

func (r *Repository) ListCustomers(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]Customer, error) {
	rows, err := tx.Query(ctx,
		`SELECT customer_id, tenant_id, name, email, created_at FROM customers WHERE tenant_id = $1 ORDER BY created_at`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing customers: %w", err)
	}
	defer rows.Close()

	var out []Customer
	for rows.Next() {
		var c Customer
		if err := rows.Scan(&c.CustomerID, &c.TenantID, &c.Name, &c.Email, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning customer: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CreateOrder relies on the customers foreign key to reject an unknown
// customer_id, rather than doing a separate existence-check SELECT
// first — one round trip instead of two, and the DB is the actual
// source of truth for referential integrity either way.
func (r *Repository) CreateOrder(ctx context.Context, tx pgx.Tx, tenantID, customerID uuid.UUID) (*Order, error) {
	var o Order
	err := tx.QueryRow(ctx,
		`INSERT INTO orders (tenant_id, customer_id, status) VALUES ($1, $2, 'placed')
		 RETURNING order_id, tenant_id, customer_id, status, created_at`,
		tenantID, customerID,
	).Scan(&o.OrderID, &o.TenantID, &o.CustomerID, &o.Status, &o.CreatedAt)
	if err != nil {
		if fkViolatesConstraint(err, "orders_customer_id_fkey") {
			return nil, ErrCustomerNotFound
		}
		return nil, fmt.Errorf("creating order: %w", err)
	}
	return &o, nil
}

// CreateOrderItem relies on the products/warehouses foreign keys the
// same way — a deliberate choice to enforce "does this product/warehouse
// actually belong to this tenant" at the database level via FK + RLS,
// rather than sales.Repository querying inventory's tables directly,
// which would cross the module boundary this project is built around.
func (r *Repository) CreateOrderItem(ctx context.Context, tx pgx.Tx, tenantID, orderID, productID, warehouseID uuid.UUID, quantity int, unitPrice float64) (*OrderItem, error) {
	var item OrderItem
	err := tx.QueryRow(ctx,
		`INSERT INTO order_items (tenant_id, order_id, product_id, warehouse_id, quantity, unit_price)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING order_item_id, tenant_id, order_id, product_id, warehouse_id, quantity, unit_price`,
		tenantID, orderID, productID, warehouseID, quantity, unitPrice,
	).Scan(&item.OrderItemID, &item.TenantID, &item.OrderID, &item.ProductID, &item.WarehouseID, &item.Quantity, &item.UnitPrice)
	if err != nil {
		switch {
		case fkViolatesConstraint(err, "order_items_product_id_fkey"):
			return nil, ErrInvalidProduct
		case fkViolatesConstraint(err, "order_items_warehouse_id_fkey"):
			return nil, ErrInvalidWarehouse
		}
		return nil, fmt.Errorf("creating order item: %w", err)
	}
	return &item, nil
}

func (r *Repository) GetOrder(ctx context.Context, tx pgx.Tx, tenantID, orderID uuid.UUID) (*Order, error) {
	var o Order
	err := tx.QueryRow(ctx,
		`SELECT order_id, tenant_id, customer_id, status, created_at
		 FROM orders WHERE tenant_id = $1 AND order_id = $2`,
		tenantID, orderID,
	).Scan(&o.OrderID, &o.TenantID, &o.CustomerID, &o.Status, &o.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("getting order: %w", err)
	}
	return &o, nil
}

func (r *Repository) ListOrderItems(ctx context.Context, tx pgx.Tx, tenantID, orderID uuid.UUID) ([]OrderItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT order_item_id, tenant_id, order_id, product_id, warehouse_id, quantity, unit_price
		 FROM order_items WHERE tenant_id = $1 AND order_id = $2`,
		tenantID, orderID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing order items: %w", err)
	}
	defer rows.Close()

	var out []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.OrderItemID, &item.TenantID, &item.OrderID, &item.ProductID, &item.WarehouseID, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, fmt.Errorf("scanning order item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// UpdateOrderStatus is used as a compensating action when fulfillment
// (the order.placed event) fails after the order itself already
// committed — see sales.Service.PlaceOrder.
func (r *Repository) UpdateOrderStatus(ctx context.Context, tx pgx.Tx, tenantID, orderID uuid.UUID, status string) error {
	_, err := tx.Exec(ctx,
		`UPDATE orders SET status = $1 WHERE tenant_id = $2 AND order_id = $3`,
		status, tenantID, orderID,
	)
	if err != nil {
		return fmt.Errorf("updating order status: %w", err)
	}
	return nil
}

// fkViolatesConstraint checks both that err is a foreign-key violation
// (SQLSTATE 23503) AND that it's specifically the named constraint —
// Postgres's default naming (<table>_<column>_fkey) is what makes the
// constraint names above predictable without hardcoding them elsewhere.
func fkViolatesConstraint(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" && pgErr.ConstraintName == constraintName
	}
	return false
}
