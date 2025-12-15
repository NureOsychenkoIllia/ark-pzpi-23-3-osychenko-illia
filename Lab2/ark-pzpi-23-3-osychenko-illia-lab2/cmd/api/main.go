//go:generate swag init -g main.go -o ../../docs

// Package main BusOptima API Server
//
//	@title			BusOptima API
//	@version		1.0
//	@description	API для системи оптимізації автобусних перевезень BusOptima
//	@termsOfService	http://swagger.io/terms/
//
//	@contact.name	API Support
//	@contact.email	support@busoptima.ua
//
//	@license.name	MIT
//	@license.url	https://opensource.org/licenses/MIT
//
//	@host		localhost:8080
//	@BasePath	/api
//
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				JWT Authorization header using the Bearer scheme. Example: "Authorization: Bearer {token}"
package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"busoptima/internal/config"

	_ "busoptima/docs"
	"busoptima/internal/handler"
	"busoptima/internal/middleware"
	"busoptima/internal/repository"
	"busoptima/internal/service"
)

func main() {
	cfg := config.Load()

	// Підключення до бази даних
	db, err := sqlx.Connect("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Налаштування пулу з'єднань
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Ініціалізація репозиторіїв
	repos := &repository.Repositories{
		User:      repository.NewUserRepository(db),
		Route:     repository.NewRouteRepository(db),
		Bus:       repository.NewBusRepository(db),
		Device:    repository.NewDeviceRepository(db),
		Trip:      repository.NewTripRepository(db),
		Event:     repository.NewPassengerEventRepository(db),
		Analytics: repository.NewAnalyticsRepository(db),
		Audit:     repository.NewAuditLogRepository(db),
	}

	// Ініціалізація сервісів
	services := &service.Services{
		Auth:      service.NewAuthService(repos.User, repos.Device, cfg.JWTSecret),
		Route:     service.NewRouteService(repos.Route, repos.Audit),
		Bus:       service.NewBusService(repos.Bus, repos.Audit),
		Trip:      service.NewTripService(repos.Trip, repos.Event, repos.Analytics, repos.Audit),
		IoT:       service.NewIoTService(repos.Device, repos.Event, repos.Trip),
		Analytics: service.NewAnalyticsService(repos.Analytics),
		Forecast:  service.NewForecastService(repos.Analytics),
	}

	// Створення Fiber додатку
	app := fiber.New(fiber.Config{
		ErrorHandler: handler.CustomErrorHandler,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin,Content-Type,Accept,Authorization",
	}))

	// Налаштування маршрутів
	setupRoutes(app, services, cfg)

	// Запуск сервера
	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(app *fiber.App, services *service.Services, cfg *config.Config) {
	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	api := app.Group("/api")

	// Публічні маршрути
	auth := api.Group("/auth")
	authHandler := handler.NewAuthHandler(services.Auth)
	auth.Post("/login", authHandler.Login)
	auth.Post("/device", authHandler.DeviceAuth)
	auth.Post("/refresh", authHandler.RefreshToken)

	// Захищені маршрути
	protected := api.Use(middleware.JWTAuth(cfg.JWTSecret))

	// IoT маршрути
	iot := protected.Group("/iot")
	iotHandler := handler.NewIoTHandler(services.IoT)
	iot.Post("/events", iotHandler.SyncEvents)
	iot.Post("/price", iotHandler.SendPriceRecommendation)
	iot.Get("/config/:tripId", iotHandler.GetTripConfig)

	// Маршрути
	routes := protected.Group("/routes")
	routeHandler := handler.NewRouteHandler(services.Route)
	routes.Get("/", middleware.RequirePermission("routes:read"), routeHandler.GetAll)
	routes.Get("/:id", middleware.RequirePermission("routes:read"), routeHandler.GetByID)
	routes.Post("/", middleware.RequirePermission("routes:write"), routeHandler.Create)
	routes.Put("/:id", middleware.RequirePermission("routes:write"), routeHandler.Update)
	routes.Delete("/:id", middleware.RequirePermission("routes:write"), routeHandler.Delete)

	// Автобуси
	buses := protected.Group("/buses")
	busHandler := handler.NewBusHandler(services.Bus)
	buses.Get("/", middleware.RequirePermission("buses:read"), busHandler.GetAll)
	buses.Get("/:id", middleware.RequirePermission("buses:read"), busHandler.GetByID)
	buses.Post("/", middleware.RequirePermission("buses:write"), busHandler.Create)
	buses.Put("/:id", middleware.RequirePermission("buses:write"), busHandler.Update)
	buses.Delete("/:id", middleware.RequirePermission("buses:write"), busHandler.Delete)

	// Рейси
	trips := protected.Group("/trips")
	tripHandler := handler.NewTripHandler(services.Trip)
	trips.Get("/", middleware.RequirePermission("routes:read"), tripHandler.GetAll)
	trips.Get("/:id", middleware.RequirePermission("routes:read"), tripHandler.GetByID)
	trips.Post("/", middleware.RequirePermission("routes:write"), tripHandler.Create)
	trips.Put("/:id", middleware.RequirePermission("routes:write"), tripHandler.Update)
	trips.Get("/:id/events", middleware.RequirePermission("routes:read"), tripHandler.GetEvents)
	trips.Get("/:id/analytics", middleware.RequirePermission("analytics:read"), tripHandler.GetAnalytics)

	// Аналітика
	analytics := protected.Group("/analytics")
	analyticsHandler := handler.NewAnalyticsHandler(services.Analytics, services.Forecast)
	analytics.Get("/dashboard", middleware.RequirePermission("analytics:read"), analyticsHandler.GetDashboard)
	analytics.Get("/forecast", middleware.RequirePermission("analytics:read"), analyticsHandler.GetForecast)
	analytics.Get("/profitability", middleware.RequirePermission("analytics:read"), analyticsHandler.GetProfitability)

	// Адміністрування
	admin := protected.Group("/admin")
	adminHandler := handler.NewAdminHandler(services.Auth)
	admin.Get("/users", middleware.RequirePermission("users:read"), adminHandler.GetUsers)
	admin.Post("/users", middleware.RequirePermission("users:write"), adminHandler.CreateUser)
	admin.Put("/users/:id", middleware.RequirePermission("users:write"), adminHandler.UpdateUser)
	admin.Put("/users/:id/role", middleware.RequirePermission("users:write"), adminHandler.UpdateUserRole)
	admin.Get("/audit-logs", middleware.RequirePermission("audit:read"), adminHandler.GetAuditLogs)
	admin.Post("/backup", middleware.RequirePermission("system:backup"), adminHandler.CreateBackup)
}
