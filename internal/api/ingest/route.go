package ingest

import (
	"github.com/gofiber/fiber/v3"
)

func RegisterRoutes(r fiber.Router) {
	r.Post("/ingest/:docID", HandleIngest)
}
