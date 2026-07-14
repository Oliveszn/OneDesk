package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/Oliveszn/OneDesk/docs"
	"github.com/Oliveszn/OneDesk/internal/auth"
	"github.com/Oliveszn/OneDesk/internal/billing"
	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/events"
	"github.com/Oliveszn/OneDesk/internal/finance"
	"github.com/Oliveszn/OneDesk/internal/inventory"
	"github.com/Oliveszn/OneDesk/internal/procurement"
	"github.com/Oliveszn/OneDesk/internal/sales"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
	"github.com/joho/godotenv"
)

// @title			OneDesk API
// @version		1.0
// @description	Multi-tenant REP system for OneDesk.
// @host			localhost:8080
// @BasePath		/v1
// @securityDefinitions.apikey BearerAuth
// @in				header
// @name			Authorization
func main() {
	godotenv.Load()
	ctx := context.Background()
	// For production, change this to slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	appDSN := mustEnv("APP_DB_DSN")
	serviceDSN := mustEnv("SERVICE_DB_DSN")
	jwtSecret := mustEnv("JWT_SECRET")

	database, err := db.New(ctx, appDSN, serviceDSN)
	if err != nil {
		log.Fatalf("connecting to database: %v", err)
	}
	defer database.Close()

	tokenService := token.NewJWTService(jwtSecret, 24*time.Hour)
	bus := events.NewBus()

	billingRepo := billing.NewRepository(database)
	billingService := billing.NewService(billingRepo, database)
	billingHandler := billing.NewHandler(billingService, logger)

	tenancyRepo := tenancy.NewRepository(database)
	tenancyService := tenancy.NewService(tenancyRepo, database)
	tenancyHandler := tenancy.NewHandler(tenancyService, logger)

	authService := auth.NewService(tenancyRepo, tokenService, billingService)
	authHandler := auth.NewHandler(authService, logger)

	inventoryRepo := inventory.NewRepository()
	inventoryService := inventory.NewService(inventoryRepo, billingService, bus, database)
	inventoryHandler := inventory.NewHandler(inventoryService, logger)

	salesRepo := sales.NewRepository()
	salesService := sales.NewService(salesRepo, billingService, bus, database)
	salesHandler := sales.NewHandler(salesService, logger)

	financeRepo := finance.NewRepository()
	financeService := finance.NewService(financeRepo, database)
	financeHandler := finance.NewHandler(financeService, logger)

	procurementRepo := procurement.NewRepository()
	procurementService := procurement.NewService(procurementRepo, bus, database)
	procurementHandler := procurement.NewHandler(procurementService, logger)

	//this the place where sales, inventory and finace is connected
	bus.Subscribe(events.TypeOrderPlaced, inventoryService.HandleOrderPlaced)
	bus.Subscribe(events.TypeOrderPlaced, financeService.HandleOrderPlaced)
	bus.Subscribe(events.TypeStockLow, procurementService.HandleStockLow)
	bus.Subscribe(events.TypePOReceived, inventoryService.HandlePOReceived)

	r := newRouter(authHandler, tenancyHandler, billingHandler, inventoryHandler, salesHandler, financeHandler, procurementHandler, tokenService)

	addr := ":8080"
	// log.Printf("listening on %s", addr)
	logger.Info("Starting server", "addr", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("missing required env var: %s", key)
	}
	return v
}
