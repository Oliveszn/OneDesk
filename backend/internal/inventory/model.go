package inventory

import (
	"time"

	"github.com/google/uuid"
)

type Warehouse struct {
	WarehouseID uuid.UUID
	TenantID    uuid.UUID
	Name        string
	CreatedAt   time.Time
}

type Product struct {
	ProductID uuid.UUID
	TenantID  uuid.UUID
	SKU       string
	Name      string
	CreatedAt time.Time
}

type StockLevel struct {
	ProductID    uuid.UUID
	WarehouseID  uuid.UUID
	TenantID     uuid.UUID
	Quantity     int
	ReorderPoint int
}
