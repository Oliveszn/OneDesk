package main

import (
	"github.com/Oliveszn/OneDesk/internal/auth"
	"github.com/Oliveszn/OneDesk/internal/middleware"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
	"github.com/go-chi/chi/v5"
)

func newRouter(authHandler *auth.Handler, tenancyHandler *tenancy.Handler, tokenService *token.JWTService) *chi.Mux {
	r := chi.NewRouter()

	// Public routes
	r.Post("/v1/tenants", authHandler.Signup)
	r.Post("/v1/auth/login", authHandler.Login)

	// Auth routes everything after this middleware has tenant_id/user_id/role available via reqctx.TenantID(ctx)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(tokenService))
		r.Get("/v1/tenants/me", tenancyHandler.Me)
	})

	return r
}
