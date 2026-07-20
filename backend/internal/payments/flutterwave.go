package payments

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const flutterwaveBaseURL = "https://api.flutterwave.com/v3"

type FlutterwaveGateway struct {
	secretKey   string
	webhookHash string // a string YOU chose, not one Flutterwave generated
	redirectURL string // where Flutterwave sends the customer back after hosted checkout
	client      *http.Client
}

func NewFlutterwaveGateway(secretKey, webhookHash, redirectURL string) *FlutterwaveGateway {
	return &FlutterwaveGateway{
		secretKey:   secretKey,
		webhookHash: webhookHash,
		redirectURL: redirectURL,
		client:      &http.Client{Timeout: 15 * time.Second},
	}
}

func (g *FlutterwaveGateway) Name() string { return "flutterwave" }

type flutterwaveEnvelope struct {
	Status  string          `json:"status"` // "success" | "error" — a string, not a bool, unlike Paystack
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (g *FlutterwaveGateway) doRequest(ctx context.Context, method, path string, body any) (*flutterwaveEnvelope, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, flutterwaveBaseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+g.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
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

	if resp.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: flutterwave returned %d", ErrGatewayDown, resp.StatusCode)
	}

	var env flutterwaveEnvelope
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("decoding flutterwave response: %w", err)
	}
	return &env, nil
}

func (g *FlutterwaveGateway) InitializeCheckout(ctx context.Context, req CheckoutRequest) (*CheckoutResult, error) {
	env, err := g.doRequest(ctx, http.MethodPost, "/payments", map[string]any{
		"tx_ref":       req.Reference,
		"amount":       req.Amount, // Flutterwave uses whole currency units, NOT subunits like Paystack's kobo
		"currency":     req.Currency,
		"redirect_url": g.redirectURL,
		"customer":     map[string]any{"email": req.Email},
	})
	if err != nil {
		return nil, err
	}
	if env.Status != "success" {
		return nil, fmt.Errorf("flutterwave: %s", env.Message)
	}

	var data struct {
		Link string `json:"link"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decoding flutterwave initialize response: %w", err)
	}

	return &CheckoutResult{CheckoutURL: data.Link, GatewayRef: req.Reference}, nil
}

func (g *FlutterwaveGateway) ChargeWithToken(ctx context.Context, req TokenChargeRequest) (*ChargeResult, error) {
	env, err := g.doRequest(ctx, http.MethodPost, "/tokenized-charges", map[string]any{
		"token":    req.AuthToken,
		"currency": req.Currency,
		"amount":   req.Amount,
		"email":    req.Email,
		"tx_ref":   req.Reference,
	})
	if err != nil {
		return nil, err
	}

	var data struct {
		Status string `json:"status"` // "successful" | "failed"
		TxRef  string `json:"tx_ref"`
	}
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return nil, fmt.Errorf("decoding flutterwave charge response: %w", err)
	}

	// getting this far means Flutterwave processed the request. "failed" here is the card being declined, not the gateway being down, never fail over on it
	if data.Status != "successful" {
		return nil, fmt.Errorf("%w: %s", ErrCardDeclined, data.Status)
	}

	return &ChargeResult{GatewayRef: data.TxRef, Status: "success"}, nil
}

// VerifyWebhook implements Flutterwave's scheme, which is NOT an HMAC —
// it's a plain string comparison against a secret hash YOU chose and
// configured in Flutterwave's dashboard
func (g *FlutterwaveGateway) VerifyWebhook(payload []byte, headers http.Header) (*WebhookEvent, error) {
	got := headers.Get("verif-hash")
	if subtle.ConstantTimeCompare([]byte(got), []byte(g.webhookHash)) != 1 {
		return nil, ErrInvalidWebhook
	}

	var body struct {
		Event string `json:"event"` // "charge.completed"
		Data  struct {
			TxRef    string  `json:"tx_ref"`
			Status   string  `json:"status"` // "successful" | "failed"
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
			Card     struct {
				Token string `json:"token"`
			} `json:"card"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return nil, fmt.Errorf("decoding flutterwave webhook payload: %w", err)
	}

	event := &WebhookEvent{
		Reference:  body.Data.TxRef,
		GatewayRef: body.Data.TxRef,
		Amount:     body.Data.Amount,
		Currency:   body.Data.Currency,
	}
	if body.Event == "charge.completed" && body.Data.Status == "successful" {
		event.Status = "success"
		event.AuthToken = body.Data.Card.Token
	} else {
		event.Status = "failed"
	}
	return event, nil
}
