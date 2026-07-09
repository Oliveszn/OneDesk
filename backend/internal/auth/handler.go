package auth

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/Oliveszn/OneDesk/internal/httputil"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/validate"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(s *Service, l *slog.Logger) *Handler {
	return &Handler{service: s, logger: l}
}

// POST /v1/tenants
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req SignupRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("signup: malformed JSON body", "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Signup(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, tenancy.ErrEmailTaken):
			h.logger.Warn("signup: email conflict", "email", req.Email)
			httputil.WriteError(w, http.StatusConflict, "email already registered")

		case errors.Is(err, validate.ErrPasswordTooShort), errors.Is(err, validate.ErrEmailRequired):
			h.logger.Warn("signup: validation failed", "error", err.Error())
			httputil.WriteError(w, http.StatusBadRequest, err.Error())

		default:
			h.logger.Error("signup: internal service failure", "error", err.Error())
			httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		}
		return
	}

	h.logger.Info("tenant registered successfully", "tenant_id", resp.TenantID, "user_id", resp.UserID)
	httputil.WriteJSON(w, http.StatusCreated, resp)
}

// POST /v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("login: malformed JSON body", "error", err.Error())
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.service.Login(r.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.logger.Warn("login: failed authentication attempt", "email", req.Email)
			httputil.WriteError(w, http.StatusUnauthorized, "invalid email or password")
			return
		}

		h.logger.Error("login: internal system failure", "error", err.Error())
		httputil.WriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.logger.Info("user logged in successfully", "user_id", resp.UserID, "tenant_id", resp.TenantID)
	httputil.WriteJSON(w, http.StatusOK, resp)
}
