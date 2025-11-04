package main

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/api/teacher"
	"ai-learn-english/internal/api/upload"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	malvus "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func main() {
	if err := config.Init("config.yaml"); err != nil {
		log.Printf("config init error: %v", err)
	}

	app := fiber.New()

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Milvus connectivity check on startup
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	cli, err := malvus.NewClient(ctx, malvus.Config{Address: "localhost:19530"})
	cancel()
	if err != nil {
		log.Printf("milvus connect error: %v", err)
	}
	cli.Close()
	log.Println("milvus ok")

	// routes
	teacher.RegisterRoutes(app)
	upload.RegisterRoutes(app)

	addr := fmt.Sprintf(":%d", config.Cfg.Server.Port)
	if err := app.Listen(addr); err != nil {
		log.Printf("server error: %v", err)
	}
}
