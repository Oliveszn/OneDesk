package main

import (
	"github.com/Oliveszn/OneDesk/internal/auth"
	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/middleware"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
	"github.com/go-chi/chi/v5"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func newRouter(authHandler *auth.Handler, tenancyHandler *tenancy.Handler, billingHandler *billing.Handler, tokenService *token.JWTService) *chi.Mux {
	r := chi.NewRouter()

	r.Get("/swagger/*", httpSwagger.Handler())

	// Public routes
	r.Post("/v1/tenants", authHandler.Signup)
	r.Post("/v1/auth/login", authHandler.Login)
	r.Get("/v1/billing/plans", billingHandler.ListPlans)

	// Auth routes everything after this middleware has tenant_id/user_id/role available via reqctx.TenantID(ctx)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(tokenService))
		r.Get("/v1/tenants/me", tenancyHandler.Me)
		r.Get("/v1/billing/usage", billingHandler.GetUsage)
	})

	return r
}
