package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"aether-node/config"
	"aether-node/internal/db"
	"aether-node/internal/ratelimit"
	"aether-node/pkg/logger"
	"aether-node/pkg/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	alarmRepo "aether-node/internal/repository/alarm"
	apikeyRepo "aether-node/internal/repository/apikey"
	authRepo "aether-node/internal/repository/auth"
	deviceRepo "aether-node/internal/repository/device"
	installationPointRepo "aether-node/internal/repository/installation_point"
	locationRepo "aether-node/internal/repository/location"
	telemetryRepo "aether-node/internal/repository/telemetry"
	thresholdRepo "aether-node/internal/repository/threshold"
	userRepo "aether-node/internal/repository/user"

	alarmSvc "aether-node/internal/service/alarm"
	apikeySvc "aether-node/internal/service/apikey"
	authSvc "aether-node/internal/service/auth"
	deviceSvc "aether-node/internal/service/device"
	installationPointSvc "aether-node/internal/service/installation_point"
	locationSvc "aether-node/internal/service/location"
	telemetrySvc "aether-node/internal/service/telemetry"
	thresholdSvc "aether-node/internal/service/threshold"
	userSvc "aether-node/internal/service/user"

	"aether-node/internal/handler"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/nats-io/nats.go"
)

func main() {
	// Load configuration from environment
	cfg := config.MustLoad()

	// Initialize structured logger
	logger.Init(cfg.Server.LogLevel, cfg.Server.LogJSON)
	log := logger.Get()

	// PostgreSQL connection pool
	pgPool, err := pgxpool.New(context.Background(), cfg.Database.DSN())
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to connect to PostgreSQL")
	}
	defer pgPool.Close()

	// Ping PostgreSQL to verify connection
	if err := pgPool.Ping(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("PostgreSQL ping failed")
	}
	log.Info().Str("component", "postgres").Msg("Connected to PostgreSQL")

	// Verify InfluxDB connectivity via HTTP ping
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pingInfluxDB(ctx, cfg.InfluxDB.URL, cfg.InfluxDB.Token); err != nil {
		log.Warn().Err(err).Str("component", "influxdb").Msg("InfluxDB ping failed (non-fatal)")
	} else {
		log.Info().Str("component", "influxdb").Msg("Connected to InfluxDB")
	}
	
	// Connect to NATS
	log.Info().Str("component", "nats").Str("url", cfg.NATS.URL).Msg("Connecting to NATS...")
	nc, err := nats.Connect(cfg.NATS.URL)
	if err != nil {
		log.Warn().Err(err).Str("component", "nats").Str("url", cfg.NATS.URL).Msg("NATS connection failed (non-fatal)")
	} else {
		log.Info().Str("component", "nats").Str("url", cfg.NATS.URL).Msg("Connected to NATS")
		defer nc.Close()
	}

	// Create db.Queries from pool for type-safe SQL operations
	queries := db.New(pgPool)

	// Initialize repositories
	userRepo := userRepo.NewUserRepository(queries)
	deviceRepo := deviceRepo.NewDeviceRepository(queries)
	locationRepo := locationRepo.NewLocationRepository(queries)
	installationPointRepo := installationPointRepo.NewInstallationPointRepository(queries)
	apiKeyRepo := apikeyRepo.NewAPIKeyRepository(queries)
	refreshTokenRepo := authRepo.NewRefreshTokenRepository(queries)
	telemetryRepo := telemetryRepo.NewTelemetryRepository(
		cfg.InfluxDB.URL, cfg.InfluxDB.Token,
		cfg.InfluxDB.Org, cfg.InfluxDB.Bucket,
	)
	thresholdRepo := thresholdRepo.NewThresholdRepository(queries)
	alarmRepo := alarmRepo.NewAlarmRepository(queries)

	// Initialize services
	userSvc := userSvc.NewUserService(userRepo)
	telemetrySvc := telemetrySvc.NewTelemetryService(telemetryRepo, deviceRepo)
	deviceSvc := deviceSvc.NewDeviceService(deviceRepo, telemetrySvc.InvalidateDeviceCache)
	locationSvc := locationSvc.NewLocationService(locationRepo)
	installationPointSvc := installationPointSvc.NewInstallationPointService(installationPointRepo)
	apiKeySvc := apikeySvc.NewAPIKeyService(apiKeyRepo, "aeth_live_pk_")
	authSvc := authSvc.NewAuthService(
		queries,
		pgPool,
		userRepo,
		refreshTokenRepo,
		cfg.JWT.Secret,
		cfg.JWT.AccessExpiry(),
		cfg.JWT.RefreshExpiry(),
	)
	thresholdSvc := thresholdSvc.NewThresholdService(thresholdRepo, nc)
	alarmSvc := alarmSvc.NewAlarmService(alarmRepo)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userSvc)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)
	locationHandler := handler.NewLocationHandler(locationSvc)
	installationPointHandler := handler.NewInstallationPointHandler(installationPointSvc)
	apiKeyHandler := handler.NewAPIKeyHandler(apiKeySvc)
	authHandler := handler.NewAuthHandler(authSvc)
	telemetryHandler := handler.NewTelemetryHandler(telemetrySvc)
	healthChecker := handler.NewHealthChecker(pgPool, cfg.InfluxDB.URL, cfg.InfluxDB.Token)
	healthHandler := handler.NewHealthHandler(healthChecker)
	thresholdHandler := handler.NewThresholdHandler(thresholdSvc)
	alarmHandler := handler.NewAlarmHandler(alarmSvc)

	// Setup Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.RequestLogger())
	e.Use(echomiddleware.Recover())
	e.Use(echomiddleware.CORSWithConfig(echomiddleware.CORSConfig{
		AllowOrigins:     cfg.Server.CORSOrigins(),
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPatch, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-Request-ID"},
		AllowCredentials: true,
	}))

	// Rate limiting: 100 requests per minute per IP (global)
	e.Use(ratelimit.Global(100, time.Minute))

	// Health check endpoints
	e.GET("/health", healthHandler.GetHealth)
	e.GET("/health/live", healthHandler.Liveness)
	e.GET("/health/ready", healthHandler.Readiness)

	// Prometheus metrics endpoint
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Auth routes (public)
	authGroup := e.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/token/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout)

	// Protected routes
	api := e.Group("")
	api.Use(echomiddleware.JWTWithConfig(echomiddleware.JWTConfig{
		SigningKey: []byte(cfg.JWT.Secret),
		TokenLookup: "header:Authorization,query:token",
		ErrorHandlerWithContext: middleware.JWTAuthErrorHandlerWithContext,
	}))

	// User routes
	api.GET("/user/me", userHandler.GetMe)
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

	// Dashboard - SSE + History routes (via RegisterRoutes)
	telemetryGroup := api.Group("/telemetry")
	telemetryHandler.RegisterRoutes(telemetryGroup)

	// Threshold routes
	api.POST("/threshold", thresholdHandler.CreateThreshold)
	api.GET("/threshold/:guid", thresholdHandler.GetThreshold)
	api.GET("/threshold/device/:device_guid", thresholdHandler.ListThresholdsByDevice)
	api.PATCH("/threshold/:guid", thresholdHandler.UpdateThreshold)
	api.DELETE("/threshold/:guid", thresholdHandler.DeleteThreshold)

	// Alarm routes
	api.GET("/alarm/active/:device_guid", alarmHandler.GetActiveAlarms)
	api.GET("/alarm/history", alarmHandler.GetAlarmHistory)
	api.GET("/alarm/stats", alarmHandler.GetAlarmStats)
	api.GET("/alarm/stream", alarmHandler.StreamAlarms)
	api.POST("/alarm/:guid/acknowledge", alarmHandler.AcknowledgeAlarm)
	api.POST("/alarm/:guid/resolve", alarmHandler.ResolveAlarm)

	// Telemetry ingestion
	e.POST("/telemetry", telemetryHandler.WriteTelemetry)

	// Graceful shutdown with connection drain
	go func() {
		if err := e.Start(cfg.Server.Address()); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Str("component", "server").Msg("Server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Str("component", "server").Msg("Shutting down server...")

	// Create shutdown context with timeout (reuse existing ctx variable)
	ctx, cancel = context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeoutDuration())
	defer cancel()

	// First: stop accepting new connections
	if err := e.Close(); err != nil {
		log.Error().Err(err).Str("component", "server").Msg("Error closing server")
	}

	// Second: wait for existing connections to drain (up to timeout)
	shutdownDone := make(chan struct{})
	go func() {
		if err := e.Shutdown(ctx); err != nil {
			log.Warn().Err(err).Str("component", "server").Msg("Graceful shutdown incomplete")
		}
		close(shutdownDone)
	}()

	// Wait for shutdown or timeout
	select {
	case <-shutdownDone:
		log.Info().Str("component", "server").Msg("Server drained successfully")
	case <-ctx.Done():
		log.Warn().Str("component", "server").Msg("Shutdown timeout reached, forcing exit")
	}

	log.Info().Str("component", "server").Msg("Server stopped")
}

// pingInfluxDB checks InfluxDB connectivity via HTTP API
func pingInfluxDB(ctx context.Context, url, token string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/health", nil)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Token "+token)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err
	}
	return nil
}
