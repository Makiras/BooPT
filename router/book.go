package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupBookRouter(router fiber.Router) {
	book := router.Group("/book")
	{
		book.Get("/", GetBookList)
		book.Get("/:id", GetBook)
		book.Post("/", UploadNewBook)
		book.Put("/:id", UpdateBook)
		book.Delete("/:id", DeleteBook)
	}
}
