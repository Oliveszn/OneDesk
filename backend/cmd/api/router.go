package main

import (
	"github.com/Oliveszn/OneDesk/internal/auth"
	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/inventory"
	"github.com/Oliveszn/OneDesk/internal/middleware"
	"github.com/Oliveszn/OneDesk/internal/sales"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
	"github.com/go-chi/chi/v5"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func newRouter(authHandler *auth.Handler, tenancyHandler *tenancy.Handler, billingHandler *billing.Handler, inventoryHandler *inventory.Handler, salesHandler *sales.Handler, tokenService *token.JWTService) *chi.Mux {
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

		r.Post("/v1/warehouses", inventoryHandler.CreateWarehouse)
		r.Get("/v1/warehouses", inventoryHandler.ListWarehouses)
		r.Post("/v1/products", inventoryHandler.CreateProduct)
		r.Get("/v1/products", inventoryHandler.ListProducts)
		r.Get("/v1/products/{productId}/stock", inventoryHandler.GetStock)
		r.Post("/v1/products/{productId}/stock/adjust", inventoryHandler.AdjustStock)

		r.Post("/v1/customers", salesHandler.CreateCustomer)
		r.Get("/v1/customers", salesHandler.ListCustomers)
		r.Post("/v1/orders", salesHandler.PlaceOrder)
		r.Get("/v1/orders/{orderId}", salesHandler.GetOrder)
	})

	return r
}
