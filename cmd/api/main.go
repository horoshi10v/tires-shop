package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/googlesheets"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/qrcode"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/storage"
	"github.com/horoshi10v/tires-shop/internal/infrastructure/telegram"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/horoshi10v/tires-shop/docs"

	"github.com/horoshi10v/tires-shop/internal/config"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
	"github.com/horoshi10v/tires-shop/internal/repository/pg"
	"github.com/horoshi10v/tires-shop/internal/service"
	"github.com/horoshi10v/tires-shop/internal/transport/http/middleware"
	v1 "github.com/horoshi10v/tires-shop/internal/transport/http/v1"
	"github.com/horoshi10v/tires-shop/pkg/database"
)

// @title           Tires Shop CRM API
// @version         1.0
// @description     This is a REST API for managing a tires and rims store/warehouse.
// @contact.name    Valentyn Khoroshylov
// @host            localhost:8083
// @BasePath        /api/v1
// @securityDefinitions.apikey RoleAuth
// @in              header
// @name            Authorization
// @description     Введите токен в формате: Bearer {token}
func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting tires-shop api", slog.String("env", cfg.Env))

	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.DB.Host,
		Port:     cfg.DB.Port,
		User:     cfg.DB.User,
		Password: cfg.DB.Password,
		DBName:   cfg.DB.Name,
		SSLMode:  cfg.DB.SSLMode,
	})
	if err != nil {
		log.Error("failed to connect to db", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("connected to postgres")

	log.Info("running migrations...")
	if err := db.AutoMigrate(
		&models.Warehouse{},
		&models.Lot{},
		&models.Order{},
		&models.OrderItem{},
		&models.OrderMessage{},
		&models.AdminNotification{},
		&models.AuditLog{},
		&models.User{},
		&models.Transfer{},
		&models.TransferItem{},
		&models.SearchSuggestionStat{},
	); err != nil {
		log.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// --- SEEDER: Create a default warehouse if none exists ---
	var warehouseCount int64
	db.Model(&models.Warehouse{}).Count(&warehouseCount)
	if warehouseCount == 0 {
		defaultWarehouse := models.Warehouse{
			Name:     "Main Kyiv Warehouse",
			Location: "Kyiv, Center",
			IsActive: true,
		}
		db.Create(&defaultWarehouse)
		log.Info("created default warehouse for testing", slog.String("warehouse_id", defaultWarehouse.ID.String()))
	} else {
		var w models.Warehouse
		db.First(&w)
		log.Info("using existing warehouse", slog.String("warehouse_id", w.ID.String()))
	}
	// ---------------------------------------------------------

	// --- 1. Init Infrastructure Services ---
	tgNotifier := telegram.NewNotifier(log, cfg.Auth.TelegramBotToken)
	tgNotifier.Start(context.Background())
	clientBotSender := telegram.NewSender(log, cfg.Auth.ClientTelegramBotToken)
	adminBotSender := telegram.NewSender(log, cfg.Auth.TelegramBotToken)

	if err := telegram.EnsureWebhook(cfg.Auth.ClientTelegramBotToken, cfg.Telegram.ClientBotWebhookURL); err != nil {
		log.Error("failed to ensure client bot webhook", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if cfg.Telegram.ClientBotWebhookURL != "" {
		log.Info("client bot webhook ensured", slog.String("url", cfg.Telegram.ClientBotWebhookURL))
	}

	qrGenerator := qrcode.NewQRGenerator()

	minioStorage, err := storage.NewMinioStorage(
		cfg.Storage.Endpoint,
		cfg.Storage.AccessKey,
		cfg.Storage.SecretKey,
		cfg.Storage.BucketName,
		cfg.Storage.PublicURL,
		cfg.Storage.UseSSL,
		log,
	)
	if err != nil {
		log.Error("failed to init minio storage", slog.String("error", err.Error()))
		os.Exit(1)
	}

	googleExporter, err := googlesheets.NewGoogleExporter(
		context.Background(),
		"google-credentials.json",
		cfg.GoogleSpreadsheetID,
	)
	if err != nil {
		log.Error("failed to init google exporter", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// DI Container
	userRepo := pg.NewUserRepository(db)
	authService := service.NewAuthService(userRepo, cfg, log)
	authHandler := v1.NewAuthHandler(authService)
	userService := service.NewUserService(userRepo, log) // Added
	userHandler := v1.NewUserHandler(userService)        // Added

	lotRepo := pg.NewLotRepository(db)
	lotService := service.NewLotService(lotRepo, log, qrGenerator)
	lotHandler := v1.NewLotHandler(lotService)
	uploadHandler := v1.NewUploadHandler(minioStorage)

	orderRepo := pg.NewOrderRepository(db)
	adminNotificationRepo := pg.NewAdminNotificationRepository(db)
	adminNotificationService := service.NewAdminNotificationService(adminNotificationRepo, userRepo, adminBotSender, log)
	orderService := service.NewOrderService(orderRepo, log, tgNotifier, clientBotSender, adminNotificationService)
	orderHandler := v1.NewOrderHandler(orderService)
	adminNotificationHandler := v1.NewAdminNotificationHandler(adminNotificationService)

	reportRepo := pg.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo, log)
	reportHandler := v1.NewReportHandler(reportService)

	auditRepo := pg.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo, log)
	auditHandler := v1.NewAuditHandler(auditService)

	transferRepo := pg.NewTransferRepository(db)
	transferService := service.NewTransferService(transferRepo, log, tgNotifier)
	transferHandler := v1.NewTransferHandler(transferService)

	warehouseRepo := pg.NewWarehouseRepository(db)
	warehouseService := service.NewWarehouseService(warehouseRepo, log)
	warehouseHandler := v1.NewWarehouseHandler(warehouseService)

	exportService := service.NewExportService(lotRepo, reportRepo, googleExporter, log)
	exportHandler := v1.NewExportHandler(exportService)

	// Router Setup
	router := gin.Default()

	// CORS Configuration
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	corsConfig.ExposeHeaders = []string{"X-Total-Count"}
	router.Use(cors.New(corsConfig))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	publicAPI := router.Group("/api/v1")
	{
		publicAPI.GET("/lots", lotHandler.ListPublic)
		publicAPI.GET("/lots/suggestions", lotHandler.ListPublicSuggestions)
		publicAPI.POST("/lots/suggestions/track", lotHandler.TrackPublicSuggestionSelection)
		publicAPI.POST("/auth/telegram", authHandler.LoginTelegram)
		publicAPI.POST("/telegram/client/webhook", orderHandler.HandleClientBotWebhook)
	}

	clientAPI := router.Group("/api/v1")
	clientAPI.Use(middleware.RequireRole(cfg.Auth.JWTSecret, "BUYER", "STAFF", "ADMIN"))
	{
		clientAPI.POST("/orders", orderHandler.Create)
		clientAPI.GET("/orders", orderHandler.ListMyOrders)
	}

	// Staff Routes
	staffAPI := router.Group("/api/v1/staff")
	staffAPI.Use(middleware.RequireRole(cfg.Auth.JWTSecret, "ADMIN", "STAFF"))
	{
		staffAPI.GET("/lots", lotHandler.ListInternal)
		staffAPI.GET("/lots/suggestions", lotHandler.ListInternalSuggestions)
		staffAPI.POST("/lots/suggestions/track", lotHandler.TrackInternalSuggestionSelection)
		staffAPI.POST("/lots", lotHandler.Create)
		staffAPI.PUT("/lots/:id", lotHandler.Update)
		staffAPI.DELETE("/lots/:id", lotHandler.Delete)
		staffAPI.GET("/lots/:id/qr", lotHandler.GetQR)
		staffAPI.GET("/orders", orderHandler.List)
		staffAPI.PATCH("/orders/:id/status", orderHandler.UpdateStatus)
		staffAPI.POST("/orders/:id/message", orderHandler.SendMessage)
		staffAPI.GET("/orders/:id/messages", orderHandler.ListMessages)
		staffAPI.GET("/transfers", transferHandler.List)
		staffAPI.GET("/transfers/:id", transferHandler.GetByID)
		staffAPI.POST("/transfers", transferHandler.Create)
		staffAPI.POST("/transfers/:id/accept", transferHandler.Accept)
		staffAPI.POST("/transfers/:id/cancel", transferHandler.Cancel)
		staffAPI.GET("/warehouses", warehouseHandler.List)
		staffAPI.POST("/lots/upload", uploadHandler.UploadPhoto)
		staffAPI.DELETE("/lots/photo", uploadHandler.DeletePhoto)
	}

	// Admin Routes
	adminAPI := router.Group("/api/v1/admin")
	adminAPI.Use(middleware.RequireRole(cfg.Auth.JWTSecret, "ADMIN"))
	{
		adminAPI.GET("/reports/pnl", reportHandler.GetPnL)
		adminAPI.POST("/warehouses", warehouseHandler.Create)
		adminAPI.PUT("/warehouses/:id", warehouseHandler.Update)
		adminAPI.DELETE("/warehouses/:id", warehouseHandler.Delete)
		adminAPI.GET("/exports/inventory", exportHandler.ExportInventory)
		adminAPI.GET("/exports/pnl", exportHandler.ExportPnL)
		adminAPI.GET("/audit-logs", auditHandler.ListAuditLogs)
		adminAPI.GET("/notifications", adminNotificationHandler.List)
		adminAPI.POST("/notifications/:id/read", adminNotificationHandler.MarkRead)

		// User Management Routes
		adminAPI.GET("/users", userHandler.ListUsers)
		adminAPI.POST("/users", userHandler.AddWorker)
		adminAPI.PUT("/users/:id/role", userHandler.UpdateRole)
		adminAPI.DELETE("/users/:id", userHandler.Delete)
	}

	router.GET("/health", func(c *gin.Context) {
		sqlDB, _ := db.DB()
		if err := sqlDB.Ping(); err != nil {
			c.JSON(500, gin.H{"status": "error", "db": "disconnected"})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "db": "connected"})
	})

	srvAddr := fmt.Sprintf(":%s", cfg.HTTPServer.Address)
	log.Info("server starting", slog.String("address", srvAddr))

	if err := router.Run(srvAddr); err != nil {
		log.Error("server error", slog.String("error", err.Error()))
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case "local":
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case "prod":
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
