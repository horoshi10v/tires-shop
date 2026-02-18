package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/horoshi10v/tires-shop/internal/config"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)
	log.Info("starting tires-shop api", slog.String("env", cfg.Env))

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"env":    cfg.Env,
		})
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
