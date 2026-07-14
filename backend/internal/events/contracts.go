package events

import (
	"errors"

	"github.com/google/uuid"
)

//Placing the contracts in a seperate file so the pubs and subs agrre instaed of imporing each oter package

const TypeOrderPlaced = "order.placed"

type OrderPlacedPayload struct {
	OrderID uuid.UUID
	Items   []OrderPlacedItem
}

type OrderPlacedItem struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int
	UnitPrice   float64
}

// ErrInsufficientStock is returned by inventory's order.placed handler
// when an order can't be fulfilled from the requested warehouse
var ErrInsufficientStock = errors.New("insufficient stock to fulfill order")

// TypeStockLow is pub by Inventory whenever a stock adjustment takes a product/warehouse below its reorder point
const TypeStockLow = "stock.low"

type StockLowPayload struct {
	ProductID         uuid.UUID
	WarehouseID       uuid.UUID
	SuggestedQuantity int
}

// TypePOReceived is pub by Procurement when a purchase order is marked received.
// Inventory subscribes to restock accordingly
const TypePOReceived = "po.received"

type POReceivedPayload struct {
	POID  uuid.UUID
	Items []POReceivedItem
}

type POReceivedItem struct {
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Quantity    int
}
