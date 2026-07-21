package billing

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/payments"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Service struct {
	repo         *Repository
	db           *db.DB
	orchestrator *payments.Orchestrator
}

func NewService(repo *Repository, d *db.DB, orchestrator *payments.Orchestrator) *Service {
	return &Service{repo: repo, db: d, orchestrator: orchestrator}
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

//PAYMENT, RECURRING BILLING, CHECKOUT, ACTIVATION PHASE

// InitiateUpgrade starts the checkout flow for a tenant moving to the Paid plan
func (s *Service) InitiateUpgrade(ctx context.Context, tenantID uuid.UUID, email, currency string) (checkoutURL string, err error) {
	paidPlan, err := s.repo.GetPlanByName(ctx, "paid")
	if err != nil {
		return "", fmt.Errorf("looking up paid plan: %w", err)
	}

	reference := uuid.NewString() // OUR reference — sent to the gateway as tx_ref/reference, echoed back in the webhook

	result, attempted, err := s.orchestrator.InitializeCheckout(ctx, payments.CheckoutRequest{
		TenantID:  tenantID,
		Email:     email,
		Amount:    paidPlan.PriceAmount,
		Currency:  currency,
		Reference: reference,
	})
	if err != nil {
		return "", fmt.Errorf("initializing checkout: %w", err)
	}

	err = s.db.WithTenant(ctx, tenantID, func(tx pgx.Tx) error {
		sub, err := s.repo.GetActiveSubscription(ctx, tx, tenantID)
		if err != nil {
			return err
		}
		if err := s.repo.SetCheckoutReference(ctx, tx, tenantID, sub.SubscriptionID, reference); err != nil {
			return err
		}
		_, err = s.repo.CreatePaymentTransaction(ctx, tx, tenantID, sub.SubscriptionID,
			result.Gateway, reference, paidPlan.PriceAmount, currency, "pending", attempted)
		return err
	})
	if err != nil {
		return "", fmt.Errorf("recording checkout: %w", err)
	}

	return result.CheckoutURL, nil
}

// HandleCheckoutWebhook processes an incoming webhook from either gateway, gatewayName comes from which endpoint the request arrived on
func (s *Service) HandleCheckoutWebhook(ctx context.Context, gatewayName string, payload []byte, headers http.Header) error {
	event, err := s.orchestrator.VerifyWebhook(gatewayName, payload, headers)
	if err != nil {
		log.Printf("billing: webhook verification failed for %s: %v", gatewayName, err)
		return fmt.Errorf("verifying webhook: %w", err)
	}

	sub, err := s.repo.GetSubscriptionByCheckoutReference(ctx, event.Reference)
	if err != nil {
		log.Printf("billing: could not resolve subscription for webhook reference %s: %v", event.Reference, err)
		return fmt.Errorf("resolving subscription for webhook: %w", err)
	}

	err = s.db.WithTenant(ctx, sub.TenantID, func(tx pgx.Tx) error {
		if event.Status != "success" {
			if err := s.repo.ClearCheckoutReference(ctx, tx, sub.TenantID, sub.SubscriptionID); err != nil {
				return err
			}
			return nil // failed payment — tenant just stays on their current plan, nothing to compensate
		}

		paidPlan, err := s.repo.GetPlanByName(ctx, "paid")
		if err != nil {
			return err
		}
		periodEnd := time.Now().AddDate(0, 1, 0) // monthly billing interval

		if err := s.repo.ActivateSubscription(ctx, tx, sub.TenantID, sub.SubscriptionID, paidPlan.PlanID, gatewayName, event.AuthToken, periodEnd); err != nil {
			return err
		}
		return s.repo.UpsertTransactionStatusByRef(ctx, tx, sub.TenantID, event.Reference, "success")
	})
	if err != nil {
		log.Printf("billing: processing webhook for subscription %s failed: %v", sub.SubscriptionID, err)
	}
	return err
}

// RunRecurringBillingOnce is called by the billing worker

// each due subscription is charged directly on ITS gateway no routing, no failover
func (s *Service) RunRecurringBillingOnce(ctx context.Context) {
	due, err := s.repo.ListDueSubscriptions(ctx, time.Now())
	if err != nil {
		log.Printf("billing: listing due subscriptions: %v", err)
		return
	}

	for _, sub := range due {
		if err := s.chargeRenewal(ctx, sub); err != nil {
			log.Printf("billing: renewal failed for subscription %s: %v", sub.SubscriptionID, err)
		}
	}
}

func (s *Service) chargeRenewal(ctx context.Context, sub Subscription) error {
	if sub.Gateway == nil || sub.GatewayAuthToken == nil {
		return fmt.Errorf("subscription %s has no gateway/token on file", sub.SubscriptionID)
	}

	return s.db.WithTenant(ctx, sub.TenantID, func(tx pgx.Tx) error {
		plan, err := s.repo.GetPlanByID(ctx, tx, sub.PlanID)
		if err != nil {
			return err
		}

		reference := uuid.NewString()
		result, chargeErr := s.orchestrator.ChargeWithToken(ctx, *sub.Gateway, payments.TokenChargeRequest{
			Email:     "", // TODO: recurring charges need the tenant's billing email on file — not yet stored anywhere; see TRADEOFFS.md candidate
			Amount:    plan.PriceAmount,
			Currency:  plan.PriceCurrency,
			AuthToken: *sub.GatewayAuthToken,
			Reference: reference,
		})

		if chargeErr != nil {
			if _, txErr := s.repo.CreatePaymentTransaction(ctx, tx, sub.TenantID, sub.SubscriptionID, *sub.Gateway, reference, plan.PriceAmount, plan.PriceCurrency, "failed", []string{*sub.Gateway}); txErr != nil {
				return txErr
			}
			return s.repo.MarkPastDue(ctx, tx, sub.TenantID, sub.SubscriptionID)
		}

		if _, err := s.repo.CreatePaymentTransaction(ctx, tx, sub.TenantID, sub.SubscriptionID, *sub.Gateway, result.GatewayRef, plan.PriceAmount, plan.PriceCurrency, "success", []string{*sub.Gateway}); err != nil {
			return err
		}
		return s.repo.ExtendPeriod(ctx, tx, sub.TenantID, sub.SubscriptionID, time.Now().AddDate(0, 1, 0))
	})
}
