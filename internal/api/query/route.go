package query

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router) {
	r.Post("/query", HandleQuery)
}
