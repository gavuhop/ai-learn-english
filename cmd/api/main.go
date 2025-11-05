package main

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/api/teacher"
	"ai-learn-english/internal/api/upload"
	"ai-learn-english/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	malvus "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func main() {
	app := fiber.New()

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Milvus connectivity check on startup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	cli, err := malvus.NewClient(ctx, malvus.Config{Address: "localhost:19530"})
	cancel()
	if err != nil {
		logger.Error(err, "milvus connect error")
	}
	cli.Close()
	logger.Info("milvus ok")

	// routes
	teacher.RegisterRoutes(app)
	upload.RegisterRoutes(app)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	if err := app.Listen(addr); err != nil {
		logger.Error(err, "server error")
	}

}
