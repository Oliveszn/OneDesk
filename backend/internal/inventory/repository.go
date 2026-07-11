package inventory

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrSKUTaken          = errors.New("sku already exists for this tenant")
	ErrProductNotFound   = errors.New("product not found")
	ErrInsufficientStock = errors.New("insufficient stock for this adjustment")
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateWarehouse(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, name string) (*Warehouse, error) {
	var w Warehouse
	err := tx.QueryRow(ctx,
		`INSERT INTO warehouses (tenant_id, name) VALUES ($1, $2)
		 RETURNING warehouse_id, tenant_id, name, created_at`,
		tenantID, name,
	).Scan(&w.WarehouseID, &w.TenantID, &w.Name, &w.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating warehouse: %w", err)
	}
	return &w, nil
}

func (r *Repository) ListWarehouses(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]Warehouse, error) {
	rows, err := tx.Query(ctx,
		`SELECT warehouse_id, tenant_id, name, created_at FROM warehouses WHERE tenant_id = $1 ORDER BY created_at`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing warehouses: %w", err)
	}
	defer rows.Close()

	var out []Warehouse
	for rows.Next() {
		var w Warehouse
		if err := rows.Scan(&w.WarehouseID, &w.TenantID, &w.Name, &w.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning warehouse: %w", err)
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

// CreateProduct assumes the caller has already consumed the tenant's product quota (billing.ConsumeEntitlement) transaction
func (r *Repository) CreateProduct(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, sku, name string) (*Product, error) {
	var p Product
	err := tx.QueryRow(ctx,
		`INSERT INTO products (tenant_id, sku, name) VALUES ($1, $2, $3)
		 RETURNING product_id, tenant_id, sku, name, created_at`,
		tenantID, sku, name,
	).Scan(&p.ProductID, &p.TenantID, &p.SKU, &p.Name, &p.CreatedAt)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrSKUTaken
		}
		return nil, fmt.Errorf("creating product: %w", err)
	}
	return &p, nil
}

func (r *Repository) GetProduct(ctx context.Context, tx pgx.Tx, tenantID, productID uuid.UUID) (*Product, error) {
	var p Product
	err := tx.QueryRow(ctx,
		`SELECT product_id, tenant_id, sku, name, created_at
		 FROM products WHERE tenant_id = $1 AND product_id = $2`,
		tenantID, productID,
	).Scan(&p.ProductID, &p.TenantID, &p.SKU, &p.Name, &p.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product: %w", err)
	}
	return &p, nil
}

func (r *Repository) ListProducts(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]Product, error) {
	rows, err := tx.Query(ctx,
		`SELECT product_id, tenant_id, sku, name, created_at FROM products WHERE tenant_id = $1 ORDER BY created_at`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	var out []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ProductID, &p.TenantID, &p.SKU, &p.Name, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning product: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// AdjustStock atomically applies delta (positive or negative) to a product's stock at a specific warehouse
func (r *Repository) AdjustStock(ctx context.Context, tx pgx.Tx, tenantID, productID, warehouseID uuid.UUID, delta int) (int, bool, error) {
	if _, err := tx.Exec(ctx,
		`INSERT INTO stock_levels (product_id, warehouse_id, tenant_id, quantity, reorder_point)
		 VALUES ($1, $2, $3, 0, 0)
		 ON CONFLICT (product_id, warehouse_id) DO NOTHING`,
		productID, warehouseID, tenantID,
	); err != nil {
		return 0, false, fmt.Errorf("ensuring stock row: %w", err)
	}

	var newQuantity int
	err := tx.QueryRow(ctx,
		`UPDATE stock_levels SET quantity = quantity + $1
		 WHERE product_id = $2 AND warehouse_id = $3 AND tenant_id = $4
		   AND quantity + $1 >= 0
		 RETURNING quantity`,
		delta, productID, warehouseID, tenantID,
	).Scan(&newQuantity)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil // would have gone negative — cap check failed
		}
		return 0, false, fmt.Errorf("adjusting stock: %w", err)
	}
	return newQuantity, true, nil
}

func (r *Repository) GetStockLevels(ctx context.Context, tx pgx.Tx, tenantID, productID uuid.UUID) ([]StockLevel, error) {
	rows, err := tx.Query(ctx,
		`SELECT product_id, warehouse_id, tenant_id, quantity, reorder_point
		 FROM stock_levels WHERE tenant_id = $1 AND product_id = $2`,
		tenantID, productID,
	)
	if err != nil {
		return nil, fmt.Errorf("getting stock levels: %w", err)
	}
	defer rows.Close()

	var out []StockLevel
	for rows.Next() {
		var s StockLevel
		if err := rows.Scan(&s.ProductID, &s.WarehouseID, &s.TenantID, &s.Quantity, &s.ReorderPoint); err != nil {
			return nil, fmt.Errorf("scanning stock level: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
