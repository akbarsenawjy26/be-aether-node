package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aether-node/internal/domain/auth"
	"aether-node/internal/domain/device"
	"aether-node/internal/domain/installation_point"
	"aether-node/internal/domain/location"
	"aether-node/internal/domain/apikey"
	"aether-node/internal/domain/user"
	"aether-node/internal/domain/telemetry"

	"aether-node/internal/repository/auth"
	"aether-node/internal/repository/device"
	"aether-node/internal/repository/location"
	"aether-node/internal/repository/installation_point"
	"aether-node/internal/repository/apikey"
	"aether-node/internal/repository/user"
	"aether-node/internal/repository/telemetry"

	"aether-node/internal/service/auth"
	"aether-node/internal/service/device"
	"aether-node/internal/service/location"
	"aether-node/internal/service/installation_point"
	"aether-node/internal/service/apikey"
	"aether-node/internal/service/user"
	"aether-node/internal/service/telemetry"

	"aether-node/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	// Initialize database connections from env vars
	dbHost := getEnv("DATABASE_HOST", "localhost")
	dbPort := getEnv("DATABASE_PORT", "5432")
	dbUser := getEnv("DATABASE_USER", "postgres")
	dbPassword := getEnv("DATABASE_PASSWORD", "postgres")
	dbName := getEnv("DATABASE_NAME", "aether_node")
	postgresDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	// PostgreSQL connection pool
	pgPool, err := pgxpool.New(context.Background(), postgresDSN)
	if err != nil {
		log.Fatalf("Unable to connect to PostgreSQL: %v", err)
	}
	defer pgPool.Close()

	// InfluxDB client
	influxURL := os.Getenv("INFLUXDB_URL")
	influxToken := os.Getenv("INFLUXDB_TOKEN")
	if influxURL == "" {
		influxURL = "http://localhost:8086"
	}
	if influxToken == "" {
		influxToken = "my-token"
	}

	influxClient := influxdb2.NewClient(influxURL, influxToken)
	defer influxClient.Close()

	// Initialize repositories
	userRepo := user_repo.NewUserRepository(pgPool)
	deviceRepo := device_repo.NewDeviceRepository(pgPool)
	locationRepo := location_repo.NewLocationRepository(pgPool)
	installationPointRepo := installation_point_repo.NewInstallationPointRepository(pgPool)
	apiKeyRepo := apikey_repo.NewAPIKeyRepository(pgPool)
	refreshTokenRepo := auth_repo.NewRefreshTokenRepository(pgPool)
	telemetryRepo := telemetry_repo.NewTelemetryRepository(influxClient, "aether-org", "telemetry")

	// Initialize services
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production"
	}

	userSvc := user_svc.NewUserService(userRepo)
	deviceSvc := device_svc.NewDeviceService(deviceRepo)
	locationSvc := location_svc.NewLocationService(locationRepo)
	installationPointSvc := installation_point_svc.NewInstallationPointService(installationPointRepo)
	apiKeySvc := apikey_svc.NewAPIKeyService(apiKeyRepo, "aeth_live_pk_")
	authSvc := auth_svc.NewAuthService(
		userRepo,
		refreshTokenRepo,
		jwtSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
	telemetrySvc := telemetry_svc.NewTelemetryService(telemetryRepo, influxClient, "aether-org", "telemetry")

	// Initialize handlers
	userHandler := user_handler.NewUserHandler(userSvc)
	deviceHandler := device_handler.NewDeviceHandler(deviceSvc)
	locationHandler := location_handler.NewLocationHandler(locationSvc)
	installationPointHandler := installation_point_handler.NewInstallationPointHandler(installationPointSvc)
	apiKeyHandler := apikey_handler.NewAPIKeyHandler(apiKeySvc)
	authHandler := auth_handler.NewAuthHandler(authSvc)
	telemetryHandler := telemetry_handler.NewTelemetryHandler(telemetrySvc)

	// Setup Echo
	e := echo.New()

	// Middleware
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORS())
	e.Use(echomiddleware.RateLimiter(nil))

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	// Auth routes (public)
	authGroup := e.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/token/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout) // Requires auth

	// Protected routes
	api := e.Group("")
	api.Use(echomiddleware.JWTWithConfig(echomiddleware.JWTConfig{
		SigningKey: []byte(jwtSecret),
	}))

	// User routes
	api.POST("/user", userHandler.CreateUser)
	api.GET("/user/:guid", userHandler.GetUser)
	api.POST("/user/list", userHandler.ListUsers)
	api.PATCH("/user/:guid", userHandler.UpdateUser)
	api.DELETE("/user/:guid", userHandler.DeleteUser)

	// Device routes
	api.POST("/device", deviceHandler.CreateDevice)
	api.GET("/device/:guid", deviceHandler.GetDevice)
	api.POST("/device/list", deviceHandler.ListDevices)
	api.PATCH("/device/:guid", deviceHandler.UpdateDevice)
	api.DELETE("/device/:guid", deviceHandler.DeleteDevice)

	// Location routes
	api.POST("/location", locationHandler.CreateLocation)
	api.GET("/location/:guid", locationHandler.GetLocation)
	api.POST("/location/list", locationHandler.ListLocations)
	api.PATCH("/location/:guid", locationHandler.UpdateLocation)
	api.DELETE("/location/:guid", locationHandler.DeleteLocation)

	// Installation Point routes
	api.POST("/installation-point", installationPointHandler.CreateInstallationPoint)
	api.GET("/installation-point/:guid", installationPointHandler.GetInstallationPoint)
	api.GET("/installation-point/:guid/relations", installationPointHandler.GetInstallationPointWithRelations)
	api.POST("/installation-point/list", installationPointHandler.ListInstallationPoints)
	api.PATCH("/installation-point/:guid", installationPointHandler.UpdateInstallationPoint)
	api.DELETE("/installation-point/:guid", installationPointHandler.DeleteInstallationPoint)

	// API Key routes
	api.POST("/apikey", apiKeyHandler.CreateAPIKey)
	api.GET("/apikey/:guid", apiKeyHandler.GetAPIKey)
	api.POST("/apikey/list", apiKeyHandler.ListAPIKeys)
	api.PATCH("/apikey/:guid", apiKeyHandler.UpdateAPIKey)
	api.DELETE("/apikey/:guid", apiKeyHandler.DeleteAPIKey)

	// Dashboard - SSE routes
	api.GET("/stream", telemetryHandler.StreamAllDevices)
	api.GET("/stream/:device-sn", telemetryHandler.StreamDevice)

	// Dashboard - History routes
	api.POST("/history/telemetry/:device-sn", telemetryHandler.GetHistory)

	// Telemetry ingestion (device API Key auth - separate middleware would be needed)
	e.POST("/telemetry", telemetryHandler.WriteTelemetry)

	// Graceful shutdown
	go func() {
		if err := e.Start(":8080"); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Println("Server stopped")
}
