package payments

import (
	"context"
	"fmt"
	"net/http"
)

// Orchestrator routes the FIRST checkout to whichever gateway suits the currency, and fails over to the other on a gateway failure
type Orchestrator struct {
	gateways map[string]PaymentGateway
}

func NewOrchestrator(paystack, flutterwave PaymentGateway) *Orchestrator {
	return &Orchestrator{
		gateways: map[string]PaymentGateway{
			paystack.Name():    paystack,
			flutterwave.Name(): flutterwave,
		},
	}
}

// routingOrder is deliberately a plain function
// the rule is simple ("NGN prefers Paystack's deeper local rails, everything else prefers Flutterwave's broader multi-currency reach")
func (o *Orchestrator) routingOrder(currency string) []string {
	if currency == "NGN" {
		return []string{"paystack", "flutterwave"}
	}
	return []string{"flutterwave", "paystack"}
}

// InitializeCheckoutResult carries which gateway actually ended up handling the checkout, the caller: billing.Service needs to persist
// this, since it determines which gateway's ChargeWithToken to call for every future renewal of this subscription
type InitializeCheckoutResult struct {
	*CheckoutResult
	Gateway string
}

func (o *Orchestrator) InitializeCheckout(ctx context.Context, req CheckoutRequest) (*InitializeCheckoutResult, []string, error) {
	order := o.routingOrder(req.Currency)
	var attempted []string
	var lastErr error

	for _, name := range order {
		gw := o.gateways[name]
		attempted = append(attempted, name)

		result, err := gw.InitializeCheckout(ctx, req)
		if err == nil {
			return &InitializeCheckoutResult{CheckoutResult: result, Gateway: name}, attempted, nil
		}
		if !IsGatewayFailure(err) {
			// A real rejection (bad currency, business-rule failure) not a reason to try the other gateway,
			return nil, attempted, err
		}
		lastErr = err
	}

	return nil, attempted, fmt.Errorf("all payment gateways failed, last error: %w", lastErr)
}

// ChargeWithToken dispatches directly to the named gateway
func (o *Orchestrator) ChargeWithToken(ctx context.Context, gatewayName string, req TokenChargeRequest) (*ChargeResult, error) {
	gw, ok := o.gateways[gatewayName]
	if !ok {
		return nil, fmt.Errorf("unknown payment gateway %q", gatewayName)
	}
	return gw.ChargeWithToken(ctx, req)
}

// VerifyWebhook dispatches to whichever gateway the webhook claims to be from
// the caller (the HTTP handler) already knows which endpoint it arrived on (/webhooks/paystack vs /webhooks/flutterwave)
func (o *Orchestrator) VerifyWebhook(gatewayName string, payload []byte, headers http.Header) (*WebhookEvent, error) {
	gw, ok := o.gateways[gatewayName]
	if !ok {
		return nil, fmt.Errorf("unknown payment gateway %q", gatewayName)
	}
	return gw.VerifyWebhook(payload, headers)
}
