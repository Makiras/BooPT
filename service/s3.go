package service

import "github.com/gofiber/fiber/v2"

func S3PresignDownloadHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).SendString("Not Implemented")
}

func S3PresignUploadHandler(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).SendString("Not Implemented")
}
