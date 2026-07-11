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
}

// ErrInsufficientStock is returned by inventory's order.placed handler
// when an order can't be fulfilled from the requested warehouse.
var ErrInsufficientStock = errors.New("insufficient stock to fulfill order")
