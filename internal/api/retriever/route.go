package retriever

import "github.com/gofiber/fiber/v3"

func RegisterRoutes(r fiber.Router) {
	grp := r.Group("/retriever")

	grp.Get("/search", HandleSearch)
}
