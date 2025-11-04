package upload

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers upload-related routes on the provided router.
func RegisterRoutes(r fiber.Router) {
	r.Post("/upload", HandleUpload)
}
