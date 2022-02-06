package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupAccountRouterPub(router fiber.Router) {
	book := router.Group("/account")
	{
		book.Post("/login", Login)
		book.Post("/register", Register)
	}
}
