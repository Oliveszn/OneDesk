package finance

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	InvoiceID uuid.UUID
	TenantID  uuid.UUID
	OrderID   uuid.UUID
	Amount    float64
	Status    string
	IssuedAt  time.Time
}

type LedgerEntry struct {
	EntryID   int64
	TenantID  uuid.UUID
	InvoiceID uuid.UUID
	EntryType string // "debit" or "credit"
	Amount    float64
	CreatedAt time.Time
}
