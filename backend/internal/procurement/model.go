package procurement

import (
	"time"

	"github.com/google/uuid"
)

type Vendor struct {
	VendorID  uuid.UUID
	TenantID  uuid.UUID
	Name      string
	CreatedAt time.Time
}

type PurchaseOrder struct {
	POID      uuid.UUID
	TenantID  uuid.UUID
	VendorID  uuid.UUID
	Status    string // suggested, sent, received, receive_issue
	CreatedAt time.Time
}

func (po *PurchaseOrder) HasVendor() bool {
	return po.VendorID != uuid.Nil
}

type POItem struct {
	POItemID    uuid.UUID
	TenantID    uuid.UUID
	POID        uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int
}
