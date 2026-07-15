package billing

import (
	"errors"
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
