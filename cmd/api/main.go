package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/config"
	"github.com/horoshi10v/tires-shop/internal/repository/models"
	"github.com/horoshi10v/tires-shop/pkg/database"
)

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
	if err := db.AutoMigrate(&models.Warehouse{}, &models.Lot{}); err != nil {
		log.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}
	log.Info("migrations applied successfully")

	router := gin.Default()

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
		log.Error("failed to start server", slog.String("error", err.Error()))
		os.Exit(1)
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
