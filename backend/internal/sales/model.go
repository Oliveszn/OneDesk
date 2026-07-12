package sales

import (
	"time"

	"github.com/google/uuid"
)

type Customer struct {
	CustomerID uuid.UUID
	TenantID   uuid.UUID
	Name       string
	Email      string
	CreatedAt  time.Time
}

type Order struct {
	OrderID    uuid.UUID
	TenantID   uuid.UUID
	CustomerID uuid.UUID
	Status     string
	CreatedAt  time.Time
}

type OrderItem struct {
	OrderItemID uuid.UUID
	TenantID    uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int
	UnitPrice   float64
}
