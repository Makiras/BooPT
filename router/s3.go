package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupS3Router(router fiber.Router) {
	s3 := router.Group("/s3")
	{
		s3.Get("/book_file", PresignedDownloadFileHandler)
		s3.Post("/book_file", PresignedUploadFileHandler)
	}
}
