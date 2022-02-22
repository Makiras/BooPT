package service

import (
	db "BooPT/database"
	m "BooPT/model"
	"BooPT/storage"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
)

func PresignedDownloadFileHandler(c *fiber.Ctx) error {
	downloadLinkId := c.Params("link_id")
	md5Hash := c.Params("md5")
	fileName := c.Params("file_name") // custom file name by frontend
	isPreview := c.Params("is_preview")

	// whether the download link exists and authenticated
	var downloadLink m.DownloadLink
	if err := db.DB.Where("id = ?", downloadLinkId).First(&downloadLink).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Download link not found",
			})
		}
		logrus.Errorf("Error getting download link: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}
	if strings.Contains(downloadLink.Link, md5Hash) == false {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Download link is not authenticated",
		})
	}

	// get minio presigned download url, valid in 2 hours
	if presignedUrl, err := storage.GetPresignedURL(downloadLink.Link, 3600*2, fileName, isPreview == "true"); err != nil {
		logrus.Errorf("Error getting presigned url: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "Presigned download url generated",
			"data":    presignedUrl,
		})
	}
}

func PresignedUploadFileHandler(c *fiber.Ctx) error {
	fileMD5Hash := c.FormValue("md5")
	fileType := c.FormValue("type")

	// avoid duplicate upload
	objectName := "unprocessed/" + fileMD5Hash + "/file." + fileType
	if err := storage.CheckObjectExistence(objectName); err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error":  "File already exists",
			"detail": objectName,
		})
	} else {
		var minioErr minio.ErrorResponse
		if errors.As(err, &minioErr); minioErr.StatusCode != 404 {
			logrus.Errorf("Error checking object existence: %#v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal server error",
			})
		}
	}

	// get minio presigned upload url, valid in 2 hours
	if presignedUrl, formData, err := storage.PostFilePresignedURL("unprocessed/"+fileMD5Hash+"/", fileType, 3600*2); err != nil {
		logrus.Errorf("Error getting presigned url: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":  "Internal server error",
			"detail": err,
		})
	} else {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "Presigned upload url generated",
			"data": fiber.Map{
				"object_name": objectName,
				"url":         presignedUrl,
				"form":        formData,
			},
		})
	}
}
