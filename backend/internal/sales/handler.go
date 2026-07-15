package sales

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/Oliveszn/OneDesk/internal/httputil"
	"github.com/Oliveszn/OneDesk/internal/reqctx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(s *Service, l *slog.Logger) *Handler {
	return &Handler{
		service: s,
		logger:  l,
	}
}

// CreateCustomer godoc
//
//	@Summary		Create a new customer profile
//	@Description	Registers a new customer account record attached to the current active tenant workspace.
//	@Tags			sales
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateCustomerRequest	true	"Customer details input parameters"
//	@Success		201		{object}	CustomerResponse
//	@Failure		400		{object}	httputil.APIError	"Invalid payload data syntax or missing required fields"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant validation credentials"
//	@Router			/v1/customers [post]
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("create_customer: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req CreateCustomerRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("create_customer: payload syntax structure decode failure", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateCustomer(r.Context(), tenantID, req)
	if err != nil {
		h.logger.Error("create_customer: storage entity persistence pipeline failed", "tenant_id", tenantID, "customer_email", req.Email, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// ListCustomers godoc
//
//	@Summary		List tenant customer accounts
//	@Description	Returns all customer profiles matching the active tenant organizational boundary.
//	@Tags			sales
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		CustomerResponse
//	@Failure		401	{object}	httputil.APIError	"Missing valid tenant validation credentials"
//	@Failure		500	{object}	httputil.APIError	"Internal service database lookup fault"
//	@Router			/v1/customers [get]
func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_customers: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListCustomers(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_customers: operational relational scan execution error", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// PlaceOrder godoc
//
//	@Summary		Place a new sales order
//	@Description	Submits a sales checkout request, consumes monthly plan order entitlement quotas, and runs transactional inventory stock deductions.
//	@Tags			sales
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateOrderRequest	true	"Order checklist containing items list details mapping"
//	@Success		201		{object}	OrderResponse
//	@Failure		400		{object}	httputil.APIError	"Invalid references or schema entity fields input data validation errors"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant validation credentials"
//	@Failure		403		{object}	httputil.APIError	"Active monthly plan order transaction cap limit exhausted"
//	@Failure		409		{object}	httputil.APIError	"Order registered but partial stock depletion triggered a fallback state"
//	@Router			/v1/orders [post]
func (h *Handler) PlaceOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("place_order: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req CreateOrderRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("place_order: JSON model parsing error bounce", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.PlaceOrder(r.Context(), tenantID, req)
	if err != nil {
		switch {
		case errors.Is(err, billing.ErrPlanLimitReached):
			h.logger.Warn("place_order: tenant monthly billing quota cap limit hit", "tenant_id", tenantID, "customer_id", req.CustomerID)
			httputil.WriteError(w, http.StatusForbidden, "order limit reached for your plan — upgrade to place more")
		case errors.Is(err, ErrCustomerNotFound):
			h.logger.Warn("place_order: customer entity query mapping resolution empty", "tenant_id", tenantID, "customer_id", req.CustomerID)
			httputil.WriteError(w, http.StatusBadRequest, "customer not found")
		case errors.Is(err, ErrInvalidProduct):
			h.logger.Warn("place_order: item reference values do not match registered products catalog records", "tenant_id", tenantID)
			httputil.WriteError(w, http.StatusBadRequest, "one or more products not found")
		case errors.Is(err, ErrInvalidWarehouse):
			h.logger.Warn("place_order: target storage space ID validation check mismatch", "tenant_id", tenantID)
			httputil.WriteError(w, http.StatusBadRequest, "one or more warehouses not found")
		case errors.Is(err, events.ErrInsufficientStock): // Update target identifier signature reference if using custom packages
			h.logger.Warn("place_order: transaction registered under stock warning safety fallback loop", "tenant_id", tenantID, "error", err.Error())
			httputil.WriteError(w, http.StatusConflict, "order placed but could not be fully stocked — marked stock_issue, "+err.Error())
		default:
			h.logger.Error("place_order: processing transaction operational runtime rollback crash", "tenant_id", tenantID, "error", err.Error())
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// GetOrder godoc
//
//	@Summary		Get details of a sales order
//	@Description	Fetches the comprehensive order composition manifest, settlement pricing math summaries, and current routing lifecycle status tags.
//	@Tags			sales
//	@Produce		json
//	@Security		BearerAuth
//	@Param			orderId	path		string	true	"System transaction reference token lookup ID (UUID)"
//	@Success		200		{object}	OrderResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed track sequence path identifier value format"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant validation credentials"
//	@Failure		404		{object}	httputil.APIError	"No transactional invoice matches the specified trace index ID"
//	@Failure		500		{object}	httputil.APIError	"Core data repository extraction execution block runtime failure"
//	@Router			/v1/orders/{orderId} [get]
func (h *Handler) GetOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("get_order: intercepted request missing active tenant identification")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	orderID, err := uuid.Parse(chi.URLParam(r, "orderId"))
	if err != nil {
		h.logger.Warn("get_order: malformed tracking route primary key layout syntax", "tenant_id", tenantID, "input_param", chi.URLParam(r, "orderId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid order id")
		return
	}

	resp, err := h.service.GetOrder(r.Context(), tenantID, orderID)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			h.logger.Warn("get_order: target sequence query returned clean empty set outcomes", "tenant_id", tenantID, "order_id", orderID)
			httputil.WriteError(w, http.StatusNotFound, "order not found")
			return
		}
		h.logger.Error("get_order: base persistence cluster extraction runtime error", "tenant_id", tenantID, "order_id", orderID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
