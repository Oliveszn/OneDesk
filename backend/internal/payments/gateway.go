package payments

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

var (
	ErrGatewayTimeout = errors.New("payment gateway timeout")
	ErrGatewayDown    = errors.New("payment gateway unavailable")
	ErrCardDeclined   = errors.New("card declined") //orchestrator must never fail over on this
	ErrInvalidWebhook = errors.New("webhook signature invalid")
)

// IsGatewayFailure distinguishes the gateway itself is down and worth trying the other gateway
// from "the customer's card was declined" and must not be retried with another gateway
func IsGatewayFailure(err error) bool {
	return errors.Is(err, ErrGatewayTimeout) || errors.Is(err, ErrGatewayDown)
}

// CheckoutRequest starts the first payment for a subscription
// Reference is OUR id (subscriptions.checkout_reference), not the gateway's
// it's how the eventual webhook gets tied back to a specific tenant's subscription.
type CheckoutRequest struct {
	TenantId  uuid.UUID
	Email     string
	Amount    float64
	Currency  string
	Reference string
}

type CheckoutResult struct {
	CheckoutURL string
	GatewayRef  string
}

// TokenChargeRequest is a recurring charge using a token obtained from a previous successful checkout.
// AuthToken's format is entirely gateway-specific (Paystack: authorization_code; Flutterwave: a card token)
// never portable between gateways
type TokenChargeRequest struct {
	Email     string
	Amount    float64
	Currency  string
	AuthToken string
	Reference string
}

type ChargeResult struct {
	GatewayRef string
	Status     string // "success" | "failed" | "pending"
}

// WebhookEvent is what both adapters normalize their gateway's payload into
// AuthToken is only populated on a successful FIRST checkout that's the moment a new recurring token is minted
type WebhookEvent struct {
	Reference  string // our checkout_reference, threading back to a subscription
	GatewayRef string
	Status     string // "success" | "failed"
	AuthToken  string // populated only on first-checkout success
	Amount     float64
	Currency   string
}

type PaymentGateway interface {
	Name() string
	InitializeCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error)
	ChargeWithToken(ctx context.Context, req TokenChargeRequest) (*ChargeResult, error)
	VerifyWebhook(payload []byte, headers http.Header) (*WebhookEvent, error)
}
