package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/Oliveszn/OneDesk/internal/auth"
	"github.com/Oliveszn/OneDesk/internal/db"
	"github.com/Oliveszn/OneDesk/internal/tenancy"
	"github.com/Oliveszn/OneDesk/internal/token"
)

func main() {
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

	tenancyRepo := tenancy.NewRepository(database)
	tenancyService := tenancy.NewService(tenancyRepo, database)
	tenancyHandler := tenancy.NewHandler(tenancyService, logger)

	authService := auth.NewService(tenancyRepo, tokenService)
	authHandler := auth.NewHandler(authService, logger)

	r := newRouter(authHandler, tenancyHandler, tokenService)

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
