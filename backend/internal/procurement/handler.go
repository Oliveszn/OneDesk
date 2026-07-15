package procurement

import (
	"errors"
	"log/slog"
	"net/http"

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

// CreateVendor godoc
//
//	@Summary		Register a new procurement vendor
//	@Description	Registers a brand new material or product vendor context under the tenant workspace ledger.
//	@Tags			procurement
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		CreateVendorRequest	true	"Vendor creation configuration payload parameters"
//	@Success		201		{object}	VendorResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed body fields or explicit repository logic restriction match"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Router			/v1/vendors [post]
func (h *Handler) CreateVendor(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("create_vendor: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req CreateVendorRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("create_vendor: malformed parsing format sequence on layout frame parameters", "tenant_id", tenantID)
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.CreateVendor(r.Context(), tenantID, req.Name)
	if err != nil {
		h.logger.Warn("create_vendor: processing execution logic baseline error mismatch", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// ListVendors godoc
//
//	@Summary		List workspace tracking vendors
//	@Description	Retrieves all vendor records assigned inside the current isolation context.
//	@Tags			procurement
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		VendorResponse
//	@Failure		401	{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Failure		500	{object}	httputil.APIError	"Internal service datastore execution error"
//	@Router			/v1/vendors [get]
func (h *Handler) ListVendors(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_vendors: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListVendors(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_vendors: transaction block compilation failed", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// ListPurchaseOrders godoc
//
//	@Summary		List purchase orders
//	@Description	Returns all tracking supply purchase orders bound to the active tenant workspace.
//	@Tags			procurement
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		PurchaseOrderResponse
//	@Failure		401	{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Failure		500	{object}	httputil.APIError	"Internal query pipeline process extraction error"
//	@Router			/v1/purchase-orders [get]
func (h *Handler) ListPurchaseOrders(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_purchase_orders: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListPurchaseOrders(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_purchase_orders: datastore log list extraction failed", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// GetPurchaseOrder godoc
//
//	@Summary		Get purchase order details
//	@Description	Fetches the layout metrics, current lifecycle stage, and structured item blocks for a unique purchase order identity token.
//	@Tags			procurement
//	@Produce		json
//	@Security		BearerAuth
//	@Param			poId	path		string	true	"System purchase order identifier tracking string (UUID)"
//	@Success		200		{object}	PurchaseOrderResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed query identity format match"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Failure		404		{object}	httputil.APIError	"Target supply purchase tracking profile is empty"
//	@Failure		500		{object}	httputil.APIError	"Internal retrieval system transaction pipeline error"
//	@Router			/v1/purchase-orders/{poId} [get]
func (h *Handler) GetPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("get_purchase_order: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	poID, err := uuid.Parse(chi.URLParam(r, "poId"))
	if err != nil {
		h.logger.Warn("get_purchase_order: routing parameter identity token parse failure", "tenant_id", tenantID, "input_param", chi.URLParam(r, "poId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid purchase order id")
		return
	}

	resp, err := h.service.GetPurchaseOrder(r.Context(), tenantID, poID)
	if err != nil {
		if errors.Is(err, ErrPONotFound) {
			h.logger.Warn("get_purchase_order: targeted profile balance index missing records", "tenant_id", tenantID, "po_id", poID)
			httputil.WriteError(w, http.StatusNotFound, "purchase order not found")
			return
		}
		h.logger.Error("get_purchase_order: extraction failure evaluating system rows", "tenant_id", tenantID, "po_id", poID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// SendPurchaseOrder godoc
//
//	@Summary		Assign vendor and send purchase order
//	@Description	Assigns a designated dispatch entity identity parameter onto a 'suggested' state template and transitions its status tracking to 'sent'.
//	@Tags			procurement
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			poId	path		string						true	"System purchase order identifier tracking string (UUID)"
//	@Param			body	body		SendPurchaseOrderRequest	true	"Identity mapping metrics for assignment execution tracking parameters"
//	@Success		200		{object}	PurchaseOrderResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed route parameters or malformed entity parameters"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Failure		404		{object}	httputil.APIError	"Target purchase order instance index mismatch"
//	@Failure		409		{object}	httputil.APIError	"Target item is not structurally isolated inside the 'suggested' status window"
//	@Failure		500		{object}	httputil.APIError	"Internal supply ledger modifications pipeline crash error"
//	@Router			/v1/purchase-orders/{poId}/send [patch]
func (h *Handler) SendPurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("send_purchase_order: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	poID, err := uuid.Parse(chi.URLParam(r, "poId"))
	if err != nil {
		h.logger.Warn("send_purchase_order: structural query route parse failure", "tenant_id", tenantID, "input_param", chi.URLParam(r, "poId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid purchase order id")
		return
	}

	var req SendPurchaseOrderRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("send_purchase_order: incoming validation object serialization failure", "tenant_id", tenantID, "po_id", poID)
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	vendorID, err := uuid.Parse(req.VendorID)
	if err != nil {
		h.logger.Warn("send_purchase_order: mapped dispatch provider field conversion token failure", "tenant_id", tenantID, "po_id", poID, "input_vendor", req.VendorID)
		httputil.WriteError(w, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	resp, err := h.service.AssignVendorAndSend(r.Context(), tenantID, poID, vendorID)
	if err != nil {
		switch {
		case errors.Is(err, ErrPONotFound):
			h.logger.Warn("send_purchase_order: procurement targeting log sequence empty row return", "tenant_id", tenantID, "po_id", poID)
			httputil.WriteError(w, http.StatusNotFound, "purchase order not found")
		case errors.Is(err, ErrVendorNotFound):
			h.logger.Warn("send_purchase_order: target assignment provider profile empty record context", "tenant_id", tenantID, "po_id", poID, "vendor_id", vendorID)
			httputil.WriteError(w, http.StatusBadRequest, "vendor not found")
		case errors.Is(err, ErrInvalidTransition):
			h.logger.Warn("send_purchase_order: execution pipeline state condition sequence block mismatch", "tenant_id", tenantID, "po_id", poID)
			httputil.WriteError(w, http.StatusConflict, "purchase order is not in 'suggested' state")
		default:
			h.logger.Error("send_purchase_order: baseline server structural framework fault context triggered", "tenant_id", tenantID, "po_id", poID, "error", err.Error())
			httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// ReceivePurchaseOrder godoc
//
//	@Summary		Fulfill and settle inventory delivery
//	@Description	Executes material receipt pipelines, updating quantities and switching the record track context state over into 'received' or tracking execution fault statuses.
//	@Tags			procurement
//	@Produce		json
//	@Security		BearerAuth
//	@Param			poId	path		string	true	"System purchase order identifier tracking string (UUID)"
//	@Template		200		{object}	PurchaseOrderResponse
//	@Failure		400		{object}	httputil.APIError	"Malformed tracking path arguments parameter"
//	@Failure		401		{object}	httputil.APIError	"Missing valid tenant credentials"
//	@Failure		409		{object}	httputil.APIError	"Fulfillment block mismatch (order not in 'sent' state or stock calculation pipeline issue)"
//	@Router			/v1/purchase-orders/{poId}/receive [post]
func (h *Handler) ReceivePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("receive_purchase_order: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	poID, err := uuid.Parse(chi.URLParam(r, "poId"))
	if err != nil {
		h.logger.Warn("receive_purchase_order: target path verification parsing format mismatch", "tenant_id", tenantID, "input_param", chi.URLParam(r, "poId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid purchase order id")
		return
	}

	resp, err := h.service.ReceivePurchaseOrder(r.Context(), tenantID, poID)
	if err != nil {
		h.logger.Warn("receive_purchase_order: processing fulfillment operation intercepted dynamic conflict mapping", "tenant_id", tenantID, "po_id", poID, "error", err.Error())
		switch {
		case errors.Is(err, ErrInvalidTransition):
			httputil.WriteError(w, http.StatusConflict, "purchase order is not in 'sent' state")
		default:
			// Covers the "received but restock failed, marked receive_issue"
			httputil.WriteError(w, http.StatusConflict, err.Error())
		}
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
