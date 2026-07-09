package tenancy

import (
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

// Me godoc
//
//	@Summary		Get current tenant profile
//	@Description	Retrieves structural information about the authenticated user's tenant account context.
//	@Tags			tenancy
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	TenantResponse
//	@Failure		401	{object}	map[string]string	"missing tenant context / unauthorized"
//	@Failure		500	{object}	map[string]string	"internal error"
//	@Router			/tenants/me [get]
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := reqctx.TenantID(r.Context())
	if !ok {
		h.logger.Warn("me: request context missing tenant_id")
		httputil.WriteError(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	resp, err := h.service.GetMyTenant(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("me: failed to retrieve tenant profile",
			"tenant_id", tenantID.String(),
			"error", err.Error(),
		)
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.logger.Debug("me: tenant fetched successfully", "tenant_id", resp.TenantID)
	httputil.WriteJSON(w, http.StatusOK, resp)
}
