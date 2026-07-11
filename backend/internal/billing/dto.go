package billing

type PlanResponse struct {
	Name              string  `json:"name" example:"free"`
	MaxUsers          *int    `json:"max_users" example:"3" extensions:"x-nullable"` // null = unlimited
	MaxProducts       *int    `json:"max_products" example:"50" extensions:"x-nullable"`
	MaxOrdersPerMonth *int    `json:"max_orders_per_month" example:"100" extensions:"x-nullable"`
	PriceAmount       float64 `json:"price_amount" example:"0.00"`
	PriceCurrency     string  `json:"price_currency" example:"NGN"`
	BillingInterval   *string `json:"billing_interval" example:"monthly" extensions:"x-nullable"`
}

type UsageResponse struct {
	PlanName     string `json:"plan_name" example:"free"`
	OrdersUsed   int    `json:"orders_used" example:"42"`
	OrdersCap    *int   `json:"orders_cap" example:"100" extensions:"x-nullable"`
	ProductsUsed int    `json:"products_used" example:"12"`
	ProductsCap  *int   `json:"products_cap" example:"50" extensions:"x-nullable"`
	UsersUsed    int    `json:"users_used" example:"2"`
	UsersCap     *int   `json:"users_cap" example:"3" extensions:"x-nullable"`
}
