package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

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
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey RoleAuth
// @in              header
// @name            X-User-Role
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

	// DI Container
	lotRepo := pg.NewLotRepository(db)
	lotService := service.NewLotService(lotRepo, log)
	lotHandler := v1.NewLotHandler(lotService)

	orderRepo := pg.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, log)
	orderHandler := v1.NewOrderHandler(orderService)

	reportRepo := pg.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo, log)
	reportHandler := v1.NewReportHandler(reportService)

	// Router Setup
	router := gin.Default()

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	publicAPI := router.Group("/api/v1")
	{
		publicAPI.GET("/lots", lotHandler.List)
		publicAPI.POST("/orders", orderHandler.Create)
	}

	// Staff Routes
	staffAPI := router.Group("/api/v1")
	staffAPI.Use(middleware.RequireRole("ADMIN", "STAFF"))
	{
		staffAPI.POST("/lots", lotHandler.Create)
		staffAPI.PATCH("/orders/:id/status", orderHandler.UpdateStatus)
	}

	// Admin Routes
	adminAPI := router.Group("/api/v1")
	adminAPI.Use(middleware.RequireRole("ADMIN"))
	{
		adminAPI.GET("/reports/pnl", reportHandler.GetPnL)
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
