package procurement

type CreateVendorRequest struct {
	Name string `json:"name" example:"Acme Industrial Supply" binding:"required"`
}

type VendorResponse struct {
	VendorID string `json:"vendor_id" example:"a1b2c3d4-5678-90ab-cdef-1234567890ab"`
	Name     string `json:"name" example:"Acme Industrial Supply"`
}

type SendPurchaseOrderRequest struct {
	VendorID string `json:"vendor_id" example:"a1b2c3d4-5678-90ab-cdef-1234567890ab" binding:"required"`
}

type POItemResponse struct {
	ProductID   string `json:"product_id" example:"p9876543-1234-abcd-ef01-234567890abc"`
	WarehouseID string `json:"warehouse_id" example:"w1112131-4151-6171-8191-011121314151"`
	Quantity    int    `json:"quantity" example:"250"`
}

type PurchaseOrderResponse struct {
	POID     string           `json:"po_id" example:"b2c3d4e5-6789-0abc-def1-234567890abc"`
	VendorID *string          `json:"vendor_id" swaggertype:"string" example:"a1b2c3d4-5678-90ab-cdef-1234567890ab"` // null until assigned
	Status   string           `json:"status" example:"suggested"`                                                    // e.g., suggested, sent, received, receive_issue
	Items    []POItemResponse `json:"items"`
}
