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

type Service struct {
	repo *Repository
	db   *db.DB
}

func NewService(repo *Repository, d *db.DB) *Service {
	return &Service{repo: repo, db: d}
}

// AssignDefaultPlan puts a newly signed-up tenant on the Free plan
func (s *Service) AssignDefaultPlan(ctx context.Context, tenantID uuid.UUID) error {
	plan, err := s.repo.GetPlanByName(ctx, "free")
	if err != nil {
		return fmt.Errorf("looking up free plan: %w", err)
	}

	err = s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		_, txErr := s.repo.CreateSubscription(ctx, tx, tenantID, plan.PlanID)
		if txErr != nil {
			return fmt.Errorf("create free default subscription: %w", txErr)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("assign default plan transaction: %w", err)
	}
	return nil
}

// CheckEntitlement answers if a tenant is currently under their cap fo a particular resource w/o mutating anything
func (s *Service) CheckEntitlement(ctx context.Context, tenantID uuid.UUID, resource Resource) error {
	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		sub, txErr := s.repo.GetActiveSubscription(ctx, tx, tenantID)
		if txErr != nil {
			return fmt.Errorf("retrieve tenant subscription context: %w", txErr)
		}

		plan, txErr := s.repo.GetPlanByID(ctx, tx, sub.PlanID)
		if txErr != nil {
			return fmt.Errorf("retrieve plan details: %w", txErr)
		}

		capValue := plan.CapFor(resource)
		if capValue == nil {
			return nil
		}

		counter, txErr := s.repo.GetUsageCounter(ctx, tx, tenantID, currentPeriodStart())
		if txErr != nil {
			return fmt.Errorf("retrieve real-time billing counters: %w", txErr)
		}

		if used(counter, resource) >= *capValue {
			return ErrPlanLimitReached
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrPlanLimitReached) {
			return err
		}
		return fmt.Errorf("entitlement check operation failed: %w", err)
	}
	return nil
}

// ConsumeEntitlement is what Inventory/Sales/ call before creating a capped resource
func (s *Service) ConsumeEntitlement(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, resource Resource) error {
	sub, err := s.repo.GetActiveSubscription(ctx, tx, tenantID)
	if err != nil {
		return fmt.Errorf("retrieve allocation sub context: %w", err)
	}
	plan, err := s.repo.GetPlanByID(ctx, tx, sub.PlanID)
	if err != nil {
		return fmt.Errorf("retrieve structural plan details: %w", err)
	}

	ok, err := s.repo.TryConsume(ctx, tx, tenantID, resource, plan.CapFor(resource), currentPeriodStart())
	if err != nil {
		return fmt.Errorf("execute counter write mutation: %w", err)
	}
	if !ok {
		return ErrPlanLimitReached
	}
	return nil
}

// GetUsage returns a tenant's current plan + usage
func (s *Service) GetUsage(ctx context.Context, tenantID uuid.UUID) (*UsageResponse, error) {
	var resp UsageResponse

	err := s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		sub, err := s.repo.GetActiveSubscription(ctx, tx, tenantID)
		if err != nil {
			return fmt.Errorf("active subscription fetch: %w", err)
		}
		plan, err := s.repo.GetPlanByID(ctx, tx, sub.PlanID)
		if err != nil {
			return fmt.Errorf("plan metadata context mapping: %w", err)
		}
		counter, err := s.repo.GetUsageCounter(ctx, tx, tenantID, currentPeriodStart())
		if err != nil {
			return fmt.Errorf("usage interval tracking counter: %w", err)
		}

		resp = UsageResponse{
			PlanName:     plan.Name,
			OrdersUsed:   counter.OrdersCount,
			OrdersCap:    plan.MaxOrdersPerMonth,
			ProductsUsed: counter.ProductsCount,
			ProductsCap:  plan.MaxProducts,
			UsersUsed:    counter.UsersCount,
			UsersCap:     plan.MaxUsers,
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("getting usage: %w", err)
	}

	return &resp, nil
}

// ListPlans returns all available plans
func (s *Service) ListPlans(ctx context.Context) ([]PlanResponse, error) {
	plans, err := s.repo.ListAllPlans(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieve plans global matrix catalog: %w", err)
	}

	resp := make([]PlanResponse, 0, len(plans))
	for _, p := range plans {
		resp = append(resp, PlanResponse{
			Name:              p.Name,
			MaxUsers:          p.MaxUsers,
			MaxProducts:       p.MaxProducts,
			MaxOrdersPerMonth: p.MaxOrdersPerMonth,
			PriceAmount:       p.PriceAmount,
			PriceCurrency:     p.PriceCurrency,
			BillingInterval:   p.BillingInterval,
		})
	}
	return resp, nil
}

// currentPeriodStart is the first of the current month, usage caps reset monthly on the calendar
func currentPeriodStart() time.Time {
	now := time.Now().UTC()
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func used(c *UsageCounter, resource Resource) int {
	switch resource {
	case ResourceOrders:
		return c.OrdersCount
	case ResourceProducts:
		return c.ProductsCount
	case ResourceUsers:
		return c.UsersCount
	default:
		return 0
	}
}
