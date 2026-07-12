package finance

import "time"

type InvoiceResponse struct {
	InvoiceID string    `json:"invoice_id" example:"e5f6g7h8-1234-5678-90ab-cdef12345678"`
	OrderID   string    `json:"order_id" example:"d3b07384-d113-4956-953e-52f01f05e3d9"`
	Amount    float64   `json:"amount" example:"1500.50"`
	Status    string    `json:"status" example:"paid"` // e.g., pending, paid, cancelled
	IssuedAt  time.Time `json:"issued_at" example:"2026-07-12T21:15:35Z"`
}
