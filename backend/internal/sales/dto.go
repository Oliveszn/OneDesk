package sales

type CreateCustomerRequest struct {
	Name  string `json:"name" example:"John Doe"`
	Email string `json:"email" example:"johndoe@example.com"`
}

type CustomerResponse struct {
	CustomerID string `json:"customer_id" example:"a4b6c8d0-1234-5678-90ab-cdef12345678"`
	Name       string `json:"name" example:"John Doe"`
	Email      string `json:"email" example:"johndoe@example.com"`
}

type OrderItemRequest struct {
	ProductID   string  `json:"product_id" example:"bc731ee0-3333-4f9e-a89c-a1d257211140"`
	WarehouseID string  `json:"warehouse_id" example:"8f076135-d858-45ec-b9cc-0320df4ee99c"`
	Quantity    int     `json:"quantity" example:"2"`
	UnitPrice   float64 `json:"unit_price" example:"25000.00"` // Client-supplied; no pricing catalog yet
}

type CreateOrderRequest struct {
	CustomerID string             `json:"customer_id" example:"a4b6c8d0-1234-5678-90ab-cdef12345678"`
	Items      []OrderItemRequest `json:"items"`
}

type OrderItemResponse struct {
	ProductID   string  `json:"product_id" example:"bc731ee0-3333-4f9e-a89c-a1d257211140"`
	WarehouseID string  `json:"warehouse_id" example:"8f076135-d858-45ec-b9cc-0320df4ee99c"`
	Quantity    int     `json:"quantity" example:"2"`
	UnitPrice   float64 `json:"unit_price" example:"25000.00"`
}

type OrderResponse struct {
	OrderID    string              `json:"order_id" example:"d3b07384-d113-4956-953e-52f01f05e3d9"`
	CustomerID string              `json:"customer_id" example:"a4b6c8d0-1234-5678-90ab-cdef12345678"`
	Status     string              `json:"status" example:"fulfilled"` // e.g., pending, fulfilled, stock_issue
	Items      []OrderItemResponse `json:"items"`
	Total      float64             `json:"total" example:"50000.00"`
}
