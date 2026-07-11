package inventory

type CreateWarehouseRequest struct {
	Name string `json:"name" example:"Lagos Central Distribution Hub"`
}

type WarehouseResponse struct {
	WarehouseID string `json:"warehouse_id" example:"8f076135-d858-45ec-b9cc-0320df4ee99c"`
	Name        string `json:"name" example:"Lagos Central Distribution Hub"`
}

type CreateProductRequest struct {
	SKU  string `json:"sku" example:"PROD-XYZ-001"`
	Name string `json:"name" example:"Ergonomic Office Chair Premium"`
}

type ProductResponse struct {
	ProductID string `json:"product_id" example:"bc731ee0-3333-4f9e-a89c-a1d257211140"`
	SKU       string `json:"sku" example:"PROD-XYZ-001"`
	Name      string `json:"name" example:"Ergonomic Office Chair Premium"`
}

type AdjustStockRequest struct {
	WarehouseID string `json:"warehouse_id" example:"8f076135-d858-45ec-b9cc-0320df4ee99c"`
	Delta       int    `json:"delta" example:"25"` // Can be negative (deductions) or positive (restocks)
}

type StockLevelResponse struct {
	ProductID    string `json:"product_id" example:"bc731ee0-3333-4f9e-a89c-a1d257211140"`
	WarehouseID  string `json:"warehouse_id" example:"8f076135-d858-45ec-b9cc-0320df4ee99c"`
	Quantity     int    `json:"quantity" example:"142"`
	ReorderPoint int    `json:"reorder_point" example:"15"`
}
