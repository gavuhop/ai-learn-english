package main

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/api/healthcheck"
	apiIngest "ai-learn-english/internal/api/ingest"
	apiRetriever "ai-learn-english/internal/api/retriever"
	"ai-learn-english/internal/api/teacher"
	"ai-learn-english/internal/api/upload"
	"ai-learn-english/pkg/logger"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	app := fiber.New(fiber.Config{
		BodyLimit:   config.Cfg.Server.BodyLimit,
		AppName:     config.Cfg.Server.AppName,
		Concurrency: config.Cfg.Server.Concurrency,
	})

	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: config.Cfg.Cors.AllowOrigins,
		AllowMethods: config.Cfg.Cors.AllowMethods,
		AllowHeaders: config.Cfg.Cors.AllowHeaders,
	}))

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// routes
	healthcheck.RegisterRoutes(app)
	apiIngest.RegisterRoutes(app)
	apiRetriever.RegisterRoutes(app)
	teacher.RegisterRoutes(app)
	upload.RegisterRoutes(app)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	logger.Info("server allow origins: %s", config.Cfg.Cors.AllowOrigins)
	logger.Info("server allow methods: %s", config.Cfg.Cors.AllowMethods)
	logger.Info("server allow headers: %s", config.Cfg.Cors.AllowHeaders)
	if err := app.Listen(addr); err != nil {
		logger.Error(err, "server error")
	}

}
