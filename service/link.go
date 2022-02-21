package service

import (
	db "BooPT/database"
	m "BooPT/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

func GetBookLink(c *fiber.Ctx) error {
	bookID := c.Params("bookID")
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	userID := uint(jwtClaims["sub"].(float64))
	stuId := int64(jwtClaims["stu_id"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// Get book from database and fetch link
	var book m.Book
	if err := db.DB.Preload("DownloadLinks", db.DB.Where("state = ?", m.BookState_Published)).First(&book, bookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No available book found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if len(book.DownloadLinks) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "No available link found"})
	}

	// Check and record user download times
	var user m.User
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	var count int64
	if timeGap, err := time.ParseDuration("-24h"); err == nil {
		if err := db.DB.Model(&m.DownloadRecord{}).Where("user_id = ? AND created_at > ?", stuId, time.Now().Add(timeGap)).Count(&count); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
		}
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if count >= user.DownloadLimit && userState <= 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Download limit reached"})
	}
	downloadRecord := m.DownloadRecord{
		UserId: stuId,
		BookId: book.ID,
	}
	db.DB.Create(&downloadRecord)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Download link fetched",
		"data":    book.DownloadLinks,
	})
}

func GetLinkList(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	offset, erro := strconv.Atoi(c.Query("offset", "0"))
	limit, errl := strconv.Atoi(c.Query("limit", "30"))
	if erro != nil || errl != nil || offset < 0 || limit < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameter"})
	}

	// Only authorized user is allowed to review new link
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "review_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to get unpub link list but he is not admin and has no \"review_book\" permission", stuId)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get link from database
	var links []m.DownloadLink
	if err := db.DB.Where("state = ?", c.Query("state", "0")).Find(&links).Offset(offset).Limit(limit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Link not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Link list fetched",
		"data":    links,
	})
}

func UploadNewBookLink(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// Only authorized user is allowed to upload new link
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "upload_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to update new link but he is not admin and has no \"upload_book\" permission", stuId)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get link from body
	var link m.DownloadLink
	if err := c.BodyParser(&link); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link"})
	}

	// Check if book id is valid
	var book m.Book
	if err := db.DB.Where(&m.Book{BID: link.BookId}).First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Book not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Check if link is valid (force using https)
	if !strings.HasPrefix(link.Link, "https://") || link.State != int64(m.BookState_Unpublished) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link"})
	}

	// Log the upload
	tx := db.DB.Begin()
	if err := tx.Create(&link).Error; err != nil {
		logrus.Errorf("Failed to create link: %v", err)
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	uploadRecord := m.UploadRecord{
		UserId:         stuId,
		BookId:         book.ID,
		DownloadLinkID: link.ID,
		UploadReason:   c.Query("reason", ""),
	}
	if err := tx.Create(&uploadRecord).Error; err != nil {
		logrus.Errorf("Failed to create link upload record: %v", err)
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if tx.Commit().Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Link uploaded",
		"data":    link,
	})
}

func UpdateLink(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// Only authorized user is allowed to review link
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "review_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to update link but he is not admin and has no \"review_book\" permission", stuId)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get link from body
	var link m.DownloadLink
	if err := c.BodyParser(&link); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link"})
	}

	// Get link from database
	var dbLink m.DownloadLink
	if err := db.DB.First(&dbLink, link.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Link not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Update link
	dbLink.BookId = m.UpdateIfNotEmpty(link.BookId, dbLink.BookId).(int64)
	dbLink.Link = m.UpdateIfNotEmpty(link.Link, dbLink.Link).(string)
	dbLink.Password = m.UpdateIfNotEmpty(link.Password, dbLink.Password).(string)
	dbLink.State = m.UpdateIfNotEmpty(link.State, dbLink.State).(int64)
	dbLink.VersionDescribe = m.UpdateIfNotEmpty(link.VersionDescribe, dbLink.VersionDescribe).(string)

	if err := db.DB.Save(&dbLink).Error; err != nil {
		logrus.Errorf("Failed to update link: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Link updated",
		"data":    dbLink,
	})
}

func DeleteLink(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// parse request body
	linkId, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid link id"})
	}

	// Check if user is admin or has "delete permission", warning: hard code 'delete_book' permission
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "delete_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to delete link %v but he is not admin and has no \"delete_book\" permission", stuId, linkId)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get link from database
	var dbLink m.DownloadLink
	if err := db.DB.First(&dbLink, linkId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Link not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Delete link
	if err := db.DB.Delete(&dbLink).Error; err != nil {
		logrus.Errorf("Failed to delete link: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Link deleted",
		"data":    nil,
	})
}
