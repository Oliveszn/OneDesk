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
