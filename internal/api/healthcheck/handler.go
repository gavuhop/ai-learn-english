package healthcheck

import (
	"ai-learn-english/config"
	"ai-learn-english/internal/database"
	"ai-learn-english/pkg/apperror"

	"context"
	"time"

	"github.com/gofiber/fiber/v3"
	malvus "github.com/milvus-io/milvus-sdk-go/v2/client"
)

func ApiHealthCheck(c fiber.Ctx) error {
	return c.SendString("ok")
}

func DatabaseHealthCheck(c fiber.Ctx) error {
	db, err := database.GetDB()
	if err != nil {
		return apperror.InternalError(config.ModuleDatabase, c, err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return apperror.InternalError(config.ModuleDatabase, c, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return apperror.InternalError(config.ModuleDatabase, c, err)
	}
	return c.SendString("ok")
}

func MilvusHealthCheck(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	cli, err := malvus.NewClient(ctx, malvus.Config{Address: config.Cfg.Milvus.Address})
	cancel()
	if err != nil {
		return apperror.InternalError(config.ModuleMilvus, c, err)
	}
	cli.Close()
	return c.SendString("ok")
}
