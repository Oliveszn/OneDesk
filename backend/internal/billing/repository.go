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
		 RETURNING subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at`,
		tenantID, planID,
	).Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("creating subscription: %w", err)
	}
	return &s, nil
}

// GetActiveSubscription reads a tenant's current subscription
func (r *Repository) GetActiveSubscription(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID) (*Subscription, error) {
	var s Subscription
	err := tx.QueryRow(ctx,
		`SELECT subscription_id, tenant_id, plan_id, status, current_period_start, current_period_end, created_at
		 FROM subscriptions
		 WHERE tenant_id = $1 AND status = 'active'
		 ORDER BY created_at DESC LIMIT 1`,
		tenantID,
	).Scan(&s.SubscriptionID, &s.TenantID, &s.PlanID, &s.Status,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt)
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
