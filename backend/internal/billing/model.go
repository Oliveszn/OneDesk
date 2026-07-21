package billing

import (
	"time"

	"github.com/google/uuid"
)

// Plan is global config  no TenantID
// A nil cap means unlimited.
type Plan struct {
	PlanID            uuid.UUID
	Name              string
	MaxUsers          *int
	MaxProducts       *int
	MaxOrdersPerMonth *int
	PriceAmount       float64
	PriceCurrency     string
	BillingInterval   *string
}

type Subscription struct {
	SubscriptionID     uuid.UUID
	TenantID           uuid.UUID
	PlanID             uuid.UUID
	Status             string
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   *time.Time
	CreatedAt          time.Time

	// payment fields, populated once a subscription has gone through a checkout

	Gateway           *string // "paystack" | "flutterwave"
	GatewayAuthToken  *string // tokenized recurring-charge credential
	CheckoutReference *string // set while a checkout is pending, cleared once resolved either way
}

// PaymentTransaction records every checkout attempt and recurring charge, successful or not
type PaymentTransaction struct {
	TransactionID     uuid.UUID
	TenantID          uuid.UUID
	SubscriptionID    uuid.UUID
	Gateway           string
	GatewayRef        string
	Amount            float64
	Currency          string
	Status            string // pending, success, failed
	AttemptedGateways []string
	CreatedAt         time.Time
}

type UsageCounter struct {
	TenantID      uuid.UUID
	PeriodStart   time.Time
	OrdersCount   int
	ProductsCount int
	UsersCount    int
}

// Resource identifies which usage counter/plan cap a check applies to
type Resource string

const (
	ResourceOrders   Resource = "orders_count"
	ResourceProducts Resource = "products_count"
	ResourceUsers    Resource = "users_count"
)

// Cafor returns the plans cap for a given resource
// we will call it to know wether tenant can create one more of something
func (p *Plan) CapFor(r Resource) *int {
	switch r {
	case ResourceOrders:
		return p.MaxOrdersPerMonth
	case ResourceProducts:
		return p.MaxProducts
	case ResourceUsers:
		return p.MaxUsers
	default:
		return nil
	}
}
