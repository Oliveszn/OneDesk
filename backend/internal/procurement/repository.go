package procurement

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrVendorNotFound    = errors.New("vendor not found")
	ErrPONotFound        = errors.New("purchase order not found")
	ErrInvalidTransition = errors.New("purchase order is not in a state that allows this action")
	ErrInvalidProduct    = errors.New("product not found or does not belong to this tenant")
	ErrInvalidWarehouse  = errors.New("warehouse not found or does not belong to this tenant")
)

// nilUUID is the sentinel used in place of NULL for vendor_id
const nilUUIDLiteral = "00000000-0000-0000-0000-000000000000"

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) CreateVendor(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, name string) (*Vendor, error) {
	var v Vendor
	err := tx.QueryRow(ctx,
		`INSERT INTO vendors (tenant_id, name) VALUES ($1, $2)
		 RETURNING vendor_id, tenant_id, name, created_at`,
		tenantID, name,
	).Scan(&v.VendorID, &v.TenantID, &v.Name, &v.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating vendor: %w", err)
	}
	return &v, nil
}

func (r *Repository) ListVendors(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]Vendor, error) {
	rows, err := tx.Query(ctx,
		`SELECT vendor_id, tenant_id, name, created_at FROM vendors WHERE tenant_id = $1 ORDER BY created_at`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing vendors: %w", err)
	}
	defer rows.Close()

	var out []Vendor
	for rows.Next() {
		var v Vendor
		if err := rows.Scan(&v.VendorID, &v.TenantID, &v.Name, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning vendor: %w", err)
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// CreatePurchaseOrder starts a PO with no vendor (NULL in the DB,
// uuid.Nil in Go) and status 'suggested
func (r *Repository) CreatePurchaseOrder(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := tx.QueryRow(ctx,
		`INSERT INTO purchase_orders (tenant_id, status) VALUES ($1, 'suggested')
		 RETURNING po_id, tenant_id, COALESCE(vendor_id, '`+nilUUIDLiteral+`'::uuid), status, created_at`,
		tenantID,
	).Scan(&po.POID, &po.TenantID, &po.VendorID, &po.Status, &po.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating purchase order: %w", err)
	}
	return &po, nil
}

func (r *Repository) CreatePOItem(ctx context.Context, tx pgx.Tx, tenantID, poID, productID, warehouseID uuid.UUID, quantity int) (*POItem, error) {
	var item POItem
	err := tx.QueryRow(ctx,
		`INSERT INTO po_items (tenant_id, po_id, product_id, warehouse_id, quantity)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING po_item_id, tenant_id, po_id, product_id, warehouse_id, quantity`,
		tenantID, poID, productID, warehouseID, quantity,
	).Scan(&item.POItemID, &item.TenantID, &item.POID, &item.ProductID, &item.WarehouseID, &item.Quantity)
	if err != nil {
		switch {
		case fkViolatesConstraint(err, "po_items_product_id_fkey"):
			return nil, ErrInvalidProduct
		case fkViolatesConstraint(err, "po_items_warehouse_id_fkey"):
			return nil, ErrInvalidWarehouse
		}
		return nil, fmt.Errorf("creating po item: %w", err)
	}
	return &item, nil
}

func (r *Repository) GetPurchaseOrder(ctx context.Context, tx pgx.Tx, tenantID, poID uuid.UUID) (*PurchaseOrder, error) {
	var po PurchaseOrder
	err := tx.QueryRow(ctx,
		`SELECT po_id, tenant_id, COALESCE(vendor_id, '`+nilUUIDLiteral+`'::uuid), status, created_at
		 FROM purchase_orders WHERE tenant_id = $1 AND po_id = $2`,
		tenantID, poID,
	).Scan(&po.POID, &po.TenantID, &po.VendorID, &po.Status, &po.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPONotFound
		}
		return nil, fmt.Errorf("getting purchase order: %w", err)
	}
	return &po, nil
}

func (r *Repository) ListPurchaseOrders(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) ([]PurchaseOrder, error) {
	rows, err := tx.Query(ctx,
		`SELECT po_id, tenant_id, COALESCE(vendor_id, '`+nilUUIDLiteral+`'::uuid), status, created_at
		 FROM purchase_orders WHERE tenant_id = $1 ORDER BY created_at DESC`,
		tenantID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing purchase orders: %w", err)
	}
	defer rows.Close()

	var out []PurchaseOrder
	for rows.Next() {
		var po PurchaseOrder
		if err := rows.Scan(&po.POID, &po.TenantID, &po.VendorID, &po.Status, &po.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning purchase order: %w", err)
		}
		out = append(out, po)
	}
	return out, rows.Err()
}

func (r *Repository) ListPOItems(ctx context.Context, tx pgx.Tx, tenantID, poID uuid.UUID) ([]POItem, error) {
	rows, err := tx.Query(ctx,
		`SELECT po_item_id, tenant_id, po_id, product_id, warehouse_id, quantity
		 FROM po_items WHERE tenant_id = $1 AND po_id = $2`,
		tenantID, poID,
	)
	if err != nil {
		return nil, fmt.Errorf("listing po items: %w", err)
	}
	defer rows.Close()

	var out []POItem
	for rows.Next() {
		var item POItem
		if err := rows.Scan(&item.POItemID, &item.TenantID, &item.POID, &item.ProductID, &item.WarehouseID, &item.Quantity); err != nil {
			return nil, fmt.Errorf("scanning po item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// AssignVendorAndSend moves a PO from 'suggested' to 'sent' with a vendor
// attached, in one statement
func (r *Repository) AssignVendorAndSend(ctx context.Context, tx pgx.Tx, tenantID, poID, vendorID uuid.UUID) error {
	tag, err := tx.Exec(ctx,
		`UPDATE purchase_orders SET vendor_id = $1, status = 'sent'
		 WHERE tenant_id = $2 AND po_id = $3 AND status = 'suggested'`,
		vendorID, tenantID, poID,
	)
	if err != nil {
		if fkViolatesConstraint(err, "purchase_orders_vendor_id_fkey") {
			return ErrVendorNotFound
		}
		return fmt.Errorf("assigning vendor: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInvalidTransition
	}
	return nil
}

// MarkReceived moves a PO from 'sent' to 'received'
func (r *Repository) MarkReceived(ctx context.Context, tx pgx.Tx, tenantID, poID uuid.UUID) error {
	tag, err := tx.Exec(ctx,
		`UPDATE purchase_orders SET status = 'received' WHERE tenant_id = $1 AND po_id = $2 AND status = 'sent'`,
		tenantID, poID,
	)
	if err != nil {
		return fmt.Errorf("marking purchase order received: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrInvalidTransition
	}
	return nil
}

// MarkReceiveIssue is the compensating action used when a PO is marked
// received but the resulting po.received event fails to restock Inventory
func (r *Repository) MarkReceiveIssue(ctx context.Context, tx pgx.Tx, tenantID, poID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE purchase_orders SET status = 'receive_issue' WHERE tenant_id = $1 AND po_id = $2`,
		tenantID, poID,
	)
	if err != nil {
		return fmt.Errorf("marking purchase order receive_issue: %w", err)
	}
	return nil
}

func fkViolatesConstraint(err error, constraintName string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23503" && pgErr.ConstraintName == constraintName
	}
	return false
}
