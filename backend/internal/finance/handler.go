package finance

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

// ListInvoices godoc
//
//	@Summary		List tenant invoice ledger records
//	@Description	Returns all issued invoices bound to the current tenant context.
//	@Tags			finance
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{array}		InvoiceResponse
//	@Failure		401	{object}	map[string]string	"Missing valid tenant credentials"
//	@Failure		500	{object}	map[string]string	"Internal service database ledger query fault"
//	@Router			/v1/invoices [get]
func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("list_invoices: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.ListInvoices(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("list_invoices: transaction log aggregation error", "tenant_id", tenantID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// GetInvoice godoc
//
//	@Summary		Get details of an invoice
//	@Description	Retrieves structural billing components, settlement balance quantities, and paid statuses for a specific invoice.
//	@Tags			finance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			invoiceId	path		string	true	"System invoice tracking index token (UUID)"
//	@Success		200			{object}	InvoiceResponse
//	@Failure		400			{object}	map[string]string	"Malformed route sequence path ID value format"
//	@Failure		401			{object}	map[string]string	"Missing valid tenant credentials"
//	@Failure		404			{object}	map[string]string	"No invoice record matches the specified balance trace token"
//	@Failure		500			{object}	map[string]string	"Internal service ledger entity retrieval database error"
//	@Router			/v1/invoices/{invoiceId} [get]
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("get_invoice: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		h.logger.Warn("get_invoice: input routing identity format parse mismatch", "tenant_id", tenantID, "input_param", chi.URLParam(r, "invoiceId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	resp, err := h.service.GetInvoice(r.Context(), tenantID, invoiceID)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			h.logger.Warn("get_invoice: target ledger entry query execution returned empty row sets", "tenant_id", tenantID, "invoice_id", invoiceID)
			httputil.WriteError(w, http.StatusNotFound, "invoice not found")
			return
		}
		h.logger.Error("get_invoice: failed processing ledger block lookups", "tenant_id", tenantID, "invoice_id", invoiceID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// PayInvoice godoc
//
//	@Summary		Process invoice settlement payment
//	@Description	Transitions an open invoice state to paid.
//	@Tags			finance
//	@Produce		json
//	@Security		BearerAuth
//	@Param			invoiceId	path		string	true	"System invoice tracking index token (UUID)"
//	@Success		200			{object}	InvoiceResponse
//	@Failure		400			{object}	map[string]string	"Malformed route sequence path ID value format"
//	@Failure		401			{object}	map[string]string	"Missing valid tenant credentials"
//	@Failure		404			{object}	map[string]string	"No invoice record matches the specified balance trace token"
//	@Failure		500			{object}	map[string]string	"Internal service balance modification execution pipeline error"
//	@Router			/v1/invoices/{invoiceId}/pay [patch]
func (h *Handler) PayInvoice(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("pay_invoice: request dropped due to missing tenant initialization")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	invoiceID, err := uuid.Parse(chi.URLParam(r, "invoiceId"))
	if err != nil {
		h.logger.Warn("pay_invoice: input routing identity format parse mismatch", "tenant_id", tenantID, "input_param", chi.URLParam(r, "invoiceId"))
		httputil.WriteError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	resp, err := h.service.PayInvoice(r.Context(), tenantID, invoiceID)
	if err != nil {
		if errors.Is(err, ErrInvoiceNotFound) {
			h.logger.Warn("pay_invoice: balance modification target sequence empty match", "tenant_id", tenantID, "invoice_id", invoiceID)
			httputil.WriteError(w, http.StatusNotFound, "invoice not found")
			return
		}
		h.logger.Error("pay_invoice: execution failed processing credit updates", "tenant_id", tenantID, "invoice_id", invoiceID, "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
