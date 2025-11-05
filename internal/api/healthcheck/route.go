package healthcheck

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router) {
	grp := r.Group("/health")

	grp.Get("/api", ApiHealthCheck)
	grp.Get("/database", DatabaseHealthCheck)
	grp.Get("/milvus", MilvusHealthCheck)
}
