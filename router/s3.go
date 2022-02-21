package router

import (
	. "BooPT/service"
	"github.com/gofiber/fiber/v2"
)

func SetupS3Router(router fiber.Router) {
	s3 := router.Group("/s3")
	{
		s3.Get("/presign", S3PresignDownloadHandler)
		s3.Post("/presign", S3PresignUploadHandler)
	}
}
