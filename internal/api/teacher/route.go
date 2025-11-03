package teacher

import (
	"github.com/gofiber/fiber/v3"
)

// RegisterRoutes registers teacher-related routes on the provided router.
func RegisterRoutes(r fiber.Router) {
	grp := r.Group("/teacher")

	grp.Get("/", GetTeacher)
}
