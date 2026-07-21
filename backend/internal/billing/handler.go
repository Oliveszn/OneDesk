package billing

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/Oliveszn/OneDesk/internal/httputil"
	"github.com/Oliveszn/OneDesk/internal/reqctx"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(s *Service, l *slog.Logger) *Handler {
	return &Handler{service: s, logger: l}
}

// ListPlans godoc
//
//	@Summary		List pricing tier catalogs
//	@Description	Returns a matrix layout of all structural subscription plan capabilities and resource limits.
//	@Tags			billing
//	@Produce		json
//	@Success		200	{array}		PlanResponse
//	@Failure		500	{object}	httputil.APIError	"internal server error context details"
//	@Router			/billing/plans [get]
func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListPlans(r.Context())
	if err != nil {
		h.logger.Error("list_plans: structural loading failure", "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, plans)
}

// GetUsage godoc
//
//	@Summary		Get tenant quota consumption
//	@Description	Fetches real-time transaction, item tracking, and staff identity usage statistics against plan cap thresholds.
//	@Tags			billing
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UsageResponse
//	@Failure		401	{object}	httputil.APIError	"missing valid tenant validation credentials"
//	@Failure		402	{object}	httputil.APIError	"plan metric consumption limitation ceiling reached"
//	@Failure		500	{object}	httputil.APIError	"internal service persistence layer failure"
//	@Router			/billing/usage [get]
func (h *Handler) GetUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("get_usage: request intercepted lacking active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	usage, err := h.service.GetUsage(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, ErrPlanLimitReached) {
			h.logger.Warn("get_usage: tenant exceeded current allocation boundary constraints", "tenant_id", tenantID.String())
			httputil.WriteError(w, http.StatusPaymentRequired, "subscription billing quota exhausted")
			return
		}

		h.logger.Error("get_usage: transactional metric generation error", "tenant_id", tenantID.String(), "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, usage)
}

// InitiateUpgrade godoc
//
//	@Summary		Start subscription upgrade checkout
//	@Description	Initiates the payment gateway session (Paystack or Flutterwave based on currency context) and returns a checkout redirect URL.
//	@Tags			billing
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		UpgradeRequest	true	"Billing profile email and desired target transaction currency"
//	@Success		200		{object}	UpgradeResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed body payload or missing email/currency parameters"
//	@Failure		401		{object}	httputil.APIError	"Missing or invalid tenant authentication token"
//	@Failure		502		{object}	httputil.APIError	"Payment gateway provider initialization failure"
//	@Router			/billing/upgrade [post]
func (h *Handler) InitiateUpgrade(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("initiate_upgrade: request dropped due to missing tenant context")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req UpgradeRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("initiate_upgrade: malformed json payload body", "tenant_id", tenantID.String())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Currency == "" {
		h.logger.Warn("initiate_upgrade: validation missing parameters", "tenant_id", tenantID.String())
		httputil.WriteError(w, http.StatusBadRequest, "email and currency are required")
		return
	}

	checkoutURL, err := h.service.InitiateUpgrade(r.Context(), tenantID, req.Email, req.Currency)
	if err != nil {
		h.logger.Error("initiate_upgrade: gateway checkout pipeline failed", "tenant_id", tenantID.String(), "currency", req.Currency, "error", err.Error())
		httputil.WriteError(w, http.StatusBadGateway, "could not start checkout: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, UpgradeResponse{CheckoutURL: checkoutURL})
}

// PaystackWebhook godoc
//
//	@Summary		Paystack payment status webhook
//	@Description	Public endpoint receiving raw transaction lifecycle events directly from Paystack. Validates payload via HMAC signature verification.
//	@Tags			billing
//	@Accept			json
//	@Produce		json
//	@Param			x-paystack-signature	header	string	true	"Cryptographic HMAC SHA512 signature computed with secret key"
//	@Success		200						"Webhook verified and transaction processed successfully"
//	@Failure		400						{object}	httputil.APIError	"Invalid payload, signature verification failure, or unknown reference"
//	@Router			/billing/webhooks/paystack [post]
func (h *Handler) PaystackWebhook(w http.ResponseWriter, r *http.Request) {
	h.handleWebhook(w, r, "paystack")
}

// FlutterwaveWebhook godoc
//
//	@Summary		Flutterwave payment status webhook
//	@Description	Public endpoint receiving transaction callback notifications from Flutterwave. Verifies secret verification header before executing internal provisioning.
//	@Tags			billing
//	@Accept			json
//	@Produce		json
//	@Param			verif-hash	header	string	true	"Secret hash verification header configured in Flutterwave dashboard"
//	@Success		200			"Webhook verified and transaction processed successfully"
//	@Failure		400			{object}	httputil.APIError	"Invalid payload, bad secret header match, or processing issue"
//	@Router			/billing/webhooks/flutterwave [post]
func (h *Handler) FlutterwaveWebhook(w http.ResponseWriter, r *http.Request) {
	h.handleWebhook(w, r, "flutterwave")
}

func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request, gateway string) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("webhook_handler: failed to read incoming body stream", "gateway", gateway, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "could not read request body")
		return
	}

	if err := h.service.HandleCheckoutWebhook(r.Context(), gateway, payload, r.Header); err != nil {
		// Deliberately a generic response regardless of the actual
		// failure reason (bad signature, unknown reference, DB error) —
		// this endpoint is public and unauthenticated by nature; it
		// shouldn't hand back diagnostic detail to whatever sent the
		// request. The real error is logged server-side inside the
		// service, not exposed here.
		h.logger.Error("webhook_handler: processing rejected", "gateway", gateway, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "webhook processing failed")
		return
	}
	h.logger.Info("webhook_handler: payload processed successfully", "gateway", gateway)
	w.WriteHeader(http.StatusOK)
}
