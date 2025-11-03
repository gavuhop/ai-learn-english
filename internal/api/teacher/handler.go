package teacher

import "github.com/gofiber/fiber/v3"

// GetTeacher returns an empty string as a minimal placeholder.
func GetTeacher(c fiber.Ctx) error {
	return c.SendString("")
}
