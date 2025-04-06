package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sweng-task/internal/config"
	"sweng-task/internal/handler"
	"sweng-task/internal/model"
	"sweng-task/internal/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func main() {
	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	log := logger.Sugar()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Infow("Configuration loaded",
		"environment", cfg.App.Environment,
		"log_level", cfg.App.LogLevel,
		"server_port", cfg.Server.Port,
	)

	// Initialize services
	lineItemService := service.NewLineItemService(log)
	adService := service.NewAdService(lineItemService, log)
	trackingEventsBuffer := 1000 // TODO: configurable from an ENV variable

	// TODO: implement tracking events storage
	discardTrackingEventsStorage := service.TrackingEventsStorageFunc(func(_ context.Context, _ []model.TrackingEvent) error { return nil })
	trackingEventsWriteTimeout := 10 * time.Second // TODO: configurable from ENV
	trackingService := service.NewTrackingService(trackingEventsBuffer, discardTrackingEventsStorage, trackingEventsWriteTimeout, log)
	go func() {
		// TODO: configurable from ENV
		chunkSize := 100
		flushEvery := 3 * time.Second
		err := trackingService.TrackingEventsWorker(ctx, chunkSize, flushEvery)
		if err != nil {
			log.Errorf("Tracking events worker stopped with an error: %v", err)

			stop() // we gracefully stop the service in case if worker is stopped
		}
	}()

	// Setup Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Ad Bidding Service",
		ReadTimeout:  cfg.Server.Timeout,
		WriteTimeout: cfg.Server.Timeout,
		IdleTimeout:  cfg.Server.Timeout,
	})

	// Register middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New())
	app.Use(cors.New())

	// Register routes
	app.Get("/health", handler.HealthCheck)

	api := app.Group("/api/v1")

	// Line Item endpoints
	lineItemHandler := handler.NewLineItemHandler(lineItemService, log)
	api.Post("/lineitems", lineItemHandler.Create)
	api.Get("/lineitems", lineItemHandler.GetAll)
	api.Get("/lineitems/:id", lineItemHandler.GetByID)

	// Ad endpoints - TO BE IMPLEMENTED BY CANDIDATE
	adHandler := handler.NewAdHandler(adService, log)
	api.Get("/ads", adHandler.GetWinningAds)

	// Tracking endpoint - TO BE IMPLEMENTED BY CANDIDATE
	trackingHandler := handler.NewTrackingHandler(trackingService, log)
	api.Post("/tracking", trackingHandler.TrackEvent)

	// Start server
	go func() {
		address := fmt.Sprintf(":%d", cfg.Server.Port)
		log.Infof("Starting server on %s", address)
		if err := app.Listen(address); err != nil {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	// Graceful shutdown
	<-ctx.Done()
	log.Info("Shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Error shutting down server: %v", err)
	}

	log.Info("Server gracefully stopped")
}
