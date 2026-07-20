package payments

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const paystackBaseURL = "https://api.paystack.co"

type PaystackGateway struct {
	secretKey string
	client    *http.Client
}

func NewPaystackGateway(secretKey string) *PaystackGateway {
	return &PaystackGateway{
		secretKey: secretKey,
		client:    &http.Client{Timeout: 15 * time.Second},
	}
}

func (g *PaystackGateway) Name() string { return "paystack" }

// paystackEnvelope is the {status, message, data} shape every Paystack API response uses
type paystackEnvelope struct {
	Status  bool            `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (g *PaystackGateway) doRequest(ctx context.Context, method, path string, body any) (*paystackEnvelope, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, paystackBaseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		// A network-level error (timeout, connection refused, DNS failure) the orchestrator should fail over
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrGatewayTimeout
		}
		return nil, fmt.Errorf("%w: %v", ErrGatewayDown, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// 5xx means Paystack's own infrastructure is failing a fail-over case
	// 4xx (bad request, invalid key, etc.) is OUR bug, not a reason to try the other gateway
	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: paystack returned %d", ErrGatewayDown, resp.StatusCode)
	}

	var env paystackEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("decoding paystack response: %w", err)
	}
	return &env, nil
}

func (g *PaystackGateway) InitializeCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error) {
	env, err := g.doRequest(ctx, http.MethodPost, "/transaction/initialize", map[string]any{
		"email":     req.Email,
		"amount":    toKobo(req.Amount),
		"currency":  req.Currency,
		"reference": req.Reference,
	})
	if err != nil {
		return nil, err
	}
	if !env.Status {
		// Request was well-formed but Paystack rejected it for a business reason (unsupported currency) not a gateway failure, don't fail over
		return nil, fmt.Errorf("paystack: %s", env.Message)
	}

	var data struct {
		AuthorizationURL string `json:"authorization_url"`
		Reference        string `json:"reference"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decoding paystack initialize response: %w", err)
	}

	return &CheckoutResult{CheckoutURL: data.AuthorizationURL, GatewayRef: data.Reference}, nil
}

func (g *PaystackGateway) ChargeWithToken(ctx context.Context, req TokenChargeRequest) (*ChargeResult, error) {
	env, err := g.doRequest(ctx, http.MethodPost, "/transaction/charge_authorization", map[string]any{
		"authorization_code": req.AuthToken,
		"email":              req.Email,
		"amount":             toKobo(req.Amount),
		"currency":           req.Currency,
		"reference":          req.Reference,
	})
	if err != nil {
		return nil, err
	}

	var data struct {
		Status    string `json:"status"` // "success" | "failed" | "abandoned": a real outcome, not a transport error
		Reference string `json:"reference"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decoding paystack charge response: %w", err)
	}

	// Getting this far means Paystack processed the request and is telling us
	// the card itself was declined
	if data.Status != "success" {
		return nil, fmt.Errorf("%w: %s", ErrCardDeclined, data.Status)
	}

	return &ChargeResult{GatewayRef: data.Reference, Status: "success"}, nil
}

// VerifyWebhook implements Paystack's signature scheme: HMAC-SHA512 of
// the raw request body, keyed with the secret key, hex-encoded, compared
// against the x-paystack-signature header
func (g *PaystackGateway) VerifyWebhook(payload []byte, headers http.Header) (*WebhookEvent, error) {
	mac := hmac.New(sha512.New, []byte(g.secretKey))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	got := headers.Get("x-paystack-signature")
	if !hmac.Equal([]byte(expected), []byte(got)) {
		return nil, ErrInvalidWebhook
	}

	var body struct {
		Event string `json:"event"`
		Data  struct {
			Reference     string  `json:"reference"`
			Status        string  `json:"status"`
			Amount        float64 `json:"amount"` // kobo
			Currency      string  `json:"currency"`
			Authorization struct {
				AuthorizationCode string `json:"authorization_code"`
			} `json:"authorization"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("decoding paystack webhook payload: %w", err)
	}

	event := &WebhookEvent{
		Reference:  body.Data.Reference,
		GatewayRef: body.Data.Reference,
		Amount:     fromKobo(body.Data.Amount),
		Currency:   body.Data.Currency,
	}
	if body.Event == "charge.success" {
		event.Status = "success"
		event.AuthToken = body.Data.Authorization.AuthorizationCode
	} else {
		event.Status = "failed"
	}
	return event, nil
}

func toKobo(naira float64) int64    { return int64(naira * 100) }
func fromKobo(kobo float64) float64 { return kobo / 100 }
