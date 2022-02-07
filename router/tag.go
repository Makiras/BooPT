package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupTagRouter(router fiber.Router) {
	tags := router.Group("/tag")
	{
		tags.Get("/", GetTagList)
		tags.Get("/:id", GetTag)
		tags.Post("/", CreateTag)
		tags.Put("/:id", UpdateTag)
		tags.Delete("/:id", DeleteTag)
	}
}
