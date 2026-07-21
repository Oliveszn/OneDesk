package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrPlanNotFound         = errors.New("plan not found")
	ErrNoActiveSubscription = errors.New("tenant has no active subscription")
	ErrPlanLimitReached     = errors.New("plan limit reached")
)

type Repository struct {
	db *db.DB
}

func NewRepository(d *db.DB) *Repository {
	return &Repository{db: d}
}

// GetPlanByName reads global plan config
func (r *Repository) GetPlanByName(ctx context.Context, name string) (*Plan, error) {
	var p Plan
	err := r.db.App.QueryRow(ctx,
		`SELECT plan_id, name, max_users, max_products, max_orders_per_month,
		        price_amount, price_currency, billing_interval
		 FROM plans WHERE name = $1`,
		name,
	).Scan(&p.PlanID, &p.Name, &p.MaxUsers, &p.MaxProducts, &p.MaxOrdersPerMonth,
		&p.PriceAmount, &p.PriceCurrency, &p.BillingInterval)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("getting plan: %w", err)
	}
	return &p, nil
}

// ListAllPlans returns every plan, cheapest first
func (r *Repository) ListAllPlans(ctx context.Context) ([]Plan, error) {
	rows, err := r.db.App.Query(ctx,
		`SELECT plan_id, name, max_users, max_products, max_orders_per_month,
		        price_amount, price_currency, billing_interval
		 FROM plans ORDER BY price_amount ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("listing plans: %w", err)
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.PlanID, &p.Name, &p.MaxUsers, &p.MaxProducts, &p.MaxOrdersPerMonth,
			&p.PriceAmount, &p.PriceCurrency, &p.BillingInterval); err != nil {
			return nil, fmt.Errorf("scanning plan: %w", err)
		}
		plans = append(plans, p)
	}
	return plans, rows.Err()
}

func (r *Repository) GetPlanByID(ctx context.Context, tx pgx.Tx, planID uuid.UUID) (*Plan, error) {
	var p Plan
	err := tx.QueryRow(ctx,
		`SELECT plan_id, name, max_users, max_products, max_orders_per_month,
		        price_amount, price_currency, billing_interval
		 FROM plans WHERE plan_id = $1`,
		planID,
	).Scan(&p.PlanID, &p.Name, &p.MaxUsers, &p.MaxProducts, &p.MaxOrdersPerMonth,
		&p.PriceAmount, &p.PriceCurrency, &p.BillingInterval)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPlanNotFound
		}
		return nil, fmt.Errorf("getting plan: %w", err)
	}
	return &p, nil
}

// CreateSubscription inserts a subscription row inside an existing tenant-scoped transaction
func (r *Repository) CreateSubscription(ctx context.Context, tx pgx.Tx, tenantID, planID uuid.UUID) (*Subscription, error) {
	var s Subscription
	err := tx.QueryRow(ctx,
		`INSERT INTO subscriptions (tenant_id, plan_id, status, current_period_start)
		 VALUES ($1, $2, 'active', NOW())
		 RETURNING subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at,
		           gateway, gateway_auth_token, checkout_reference`,
		tenantID, planID,
	).Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
		&s.Gateway, &s.GatewayAuthToken, &s.CheckoutReference)
	if err != nil {
		return nil, fmt.Errorf("creating subscription: %w", err)
	}
	return &s, nil
}

// GetActiveSubscription reads a tenant's current subscription
func (r *Repository) GetActiveSubscription(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (*Subscription, error) {
	var s Subscription
	err := tx.QueryRow(ctx,
		`SELECT subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at,
		        gateway, gateway_auth_token, checkout_reference
		 FROM subscriptions
		 WHERE tenant_id = $1 AND status = 'active'
		 ORDER BY created_at DESC LIMIT 1`,
		tenantID,
	).Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
		&s.Gateway, &s.GatewayAuthToken, &s.CheckoutReference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActiveSubscription
		}
		return nil, fmt.Errorf("getting active subscription: %w", err)
	}
	return &s, nil
}

// GetUsageCounter reads the current usage row
func (r *Repository) GetUsageCounter(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, periodStart time.Time) (*UsageCounter, error) {
	var c UsageCounter
	err := tx.QueryRow(ctx,
		`SELECT tenant_id, period_start, orders_count, products_count, users_count
		 FROM usage_counters WHERE tenant_id = $1 AND period_start = $2`,
		tenantID, periodStart,
	).Scan(&c.TenantID, &c.PeriodStart, &c.OrdersCount, &c.ProductsCount, &c.UsersCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &UsageCounter{TenantID: tenantID, PeriodStart: periodStart}, nil
		}
		return nil, fmt.Errorf("getting usage counter: %w", err)
	}
	return &c, nil
}

func (r *Repository) TryConsume(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, resource Resource, cap *int, periodStart time.Time) (bool, error) {
	column := string(resource)

	if _, err := tx.Exec(ctx,
		`INSERT INTO usage_counters (tenant_id, period_start) VALUES ($1, $2)
		 ON CONFLICT (tenant_id, period_start) DO NOTHING`,
		tenantID, periodStart,
	); err != nil {
		return false, fmt.Errorf("ensuring usage counter row: %w", err)
	}

	query := fmt.Sprintf(
		`UPDATE usage_counters SET %s = %s + 1
		 WHERE tenant_id = $1 AND period_start = $2 AND ($3::int IS NULL OR %s < $3)
		 RETURNING %s`,
		column, column, column, column,
	)

	var newCount int
	err := tx.QueryRow(ctx, query, tenantID, periodStart, cap).Scan(&newCount)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil // cap reached zero rows matched the WHERE clause
		}
		return false, fmt.Errorf("consuming usage: %w", err)
	}
	return true, nil
}

//PAYMENT, RECURRING BILLING, CHECKOUT, ACTIVATION PHASE

// SetCheckoutReference marks a subscription as having a payment in flight set at checkout initiation, cleared once the webhook resolves it either way
func (r *Repository) SetCheckoutReference(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID uuid.UUID, reference string) error {
	_, err := tx.Exec(ctx,
		`UPDATE subscriptions SET checkout_reference = $1 WHERE tenant_id = $2 AND subscription_id = $3`,
		reference, tenantID, subscriptionID,
	)
	if err != nil {
		return fmt.Errorf("setting checkout reference: %w", err)
	}
	return nil
}

// intentionally the only non-tenant-scoped query in the billing repository, used solely to identify the subscription from the checkout reference before normal tenant-scoped operations continue
func (r *Repository) GetSubscriptionByCheckoutReference(ctx context.Context, reference string) (*Subscription, error) {
	var s Subscription
	err := r.db.Service.QueryRow(ctx,
		`SELECT subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at,
		        gateway, gateway_auth_token, checkout_reference
		 FROM subscriptions WHERE checkout_reference = $1`,
		reference,
	).Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
		&s.Gateway, &s.GatewayAuthToken, &s.CheckoutReference)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNoActiveSubscription
		}
		return nil, fmt.Errorf("looking up subscription by checkout reference: %w", err)
	}
	return &s, nil
}

// ActivateSubscription is called once a webhook confirms a successful checkout and moves the tenant onto the paid plan
func (r *Repository) ActivateSubscription(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID, planID uuid.UUID, gateway, authToken string, periodEnd time.Time) error {
	_, err := tx.Exec(ctx,
		`UPDATE subscriptions
		 SET plan_id = $1, gateway = $2, gateway_auth_token = $3,
		     current_period_start = NOW(), current_period_end = $4,
		     checkout_reference = NULL, status = 'active'
		 WHERE tenant_id = $5 AND subscription_id = $6`,
		planID, gateway, authToken, periodEnd, tenantID, subscriptionID,
	)
	if err != nil {
		return fmt.Errorf("activating subscription: %w", err)
	}
	return nil
}

// ClearCheckoutReference is used when a checkout webhook reports failure, the tenant stays on whatever plan they were already on
func (r *Repository) ClearCheckoutReference(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE subscriptions SET checkout_reference = NULL WHERE tenant_id = $1 AND subscription_id = $2`,
		tenantID, subscriptionID,
	)
	if err != nil {
		return fmt.Errorf("clearing checkout reference: %w", err)
	}
	return nil
}

// MarkPastDue is used by the recurring billing worker when a renewal charge fails
func (r *Repository) MarkPastDue(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID uuid.UUID) error {
	_, err := tx.Exec(ctx,
		`UPDATE subscriptions SET status = 'past_due' WHERE tenant_id = $1 AND subscription_id = $2`,
		tenantID, subscriptionID,
	)
	if err != nil {
		return fmt.Errorf("marking subscription past_due: %w", err)
	}
	return nil
}

// ExtendPeriod is used after a successful recurring charge.
func (r *Repository) ExtendPeriod(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID uuid.UUID, newPeriodEnd time.Time) error {
	_, err := tx.Exec(ctx,
		`UPDATE subscriptions SET current_period_end = $1 WHERE tenant_id = $2 AND subscription_id = $3`,
		newPeriodEnd, tenantID, subscriptionID,
	)
	if err != nil {
		return fmt.Errorf("extending subscription period: %w", err)
	}
	return nil
}

// ListDueSubscriptions is a system-level sweep across every tenant's subscriptions called by the billing worker
func (r *Repository) ListDueSubscriptions(ctx context.Context, before time.Time) ([]Subscription, error) {
	rows, err := r.db.Service.Query(ctx,
		`SELECT subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at,
		        gateway, gateway_auth_token, checkout_reference
		 FROM subscriptions
		 WHERE status = 'active' AND gateway IS NOT NULL AND current_period_end IS NOT NULL AND current_period_end <= $1`,
		before,
	)
	if err != nil {
		return nil, fmt.Errorf("listing due subscriptions: %w", err)
	}
	defer rows.Close()

	var out []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
			&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt,
			&s.Gateway, &s.GatewayAuthToken, &s.CheckoutReference); err != nil {
			return nil, fmt.Errorf("scanning subscription: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// CreatePaymentTransaction records a checkout attempt or recurring charge attempt. gatewayRef must be unique
func (r *Repository) CreatePaymentTransaction(ctx context.Context, tx pgx.Tx, tenantID, subscriptionID uuid.UUID, gateway, gatewayRef string, amount float64, currency, status string, attemptedGateways []string) (*PaymentTransaction, error) {
	var t PaymentTransaction
	err := tx.QueryRow(ctx,
		`INSERT INTO payment_transactions (tenant_id, subscription_id, gateway, gateway_ref, amount, currency, status, attempted_gateways)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING transaction_id, tenant_id, subscription_id, gateway, gateway_ref, amount, currency, status, attempted_gateways, created_at`,
		tenantID, subscriptionID, gateway, gatewayRef, amount, currency, status, attemptedGateways,
	).Scan(&t.TransactionID, &t.TenantID, &t.SubscriptionID, &t.Gateway, &t.GatewayRef,
		&t.Amount, &t.Currency, &t.Status, &t.AttemptedGateways, &t.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating payment transaction: %w", err)
	}
	return &t, nil
}

// UpsertTransactionStatusByRef is how a webhook updates a transactions final status, inside the SAME tenant-scoped transaction as
// ActivateSubscription (both called from within HandleCheckoutWebhook's db.WithTenant)
func (r *Repository) UpsertTransactionStatusByRef(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, gatewayRef, status string) error {
	_, err := tx.Exec(ctx,
		`UPDATE payment_transactions SET status = $1 WHERE tenant_id = $2 AND gateway_ref = $3`,
		status, tenantID, gatewayRef,
	)
	if err != nil {
		return fmt.Errorf("updating payment transaction status: %w", err)
	}
	return nil
}
