package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupLinkRouter(router fiber.Router) {
	book := router.Group("/book")
	link := router.Group("/link")
	{
		book.Get("/:bookID/link", GetBookLink)
		book.Post("/:bookID/link", UploadNewBookLink)

		link.Get("/", GetLinkList)
		link.Put("/:id", UpdateLink)
		link.Delete("/:id", DeleteLink)
	}
}
