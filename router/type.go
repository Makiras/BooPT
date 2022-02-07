package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupTypeRouter(router fiber.Router) {
	type_ := router.Group("/type")
	{
		type_.Get("/", GetAllType)
		type_.Get("/:id", GetType)
		type_.Post("/", CreateType)
		type_.Put("/:id", UpdateType)
		type_.Delete("/:id", DeleteType)
	}
}
