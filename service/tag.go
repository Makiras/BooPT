package service

import (
	db "BooPT/database"
	m "BooPT/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"strconv"
)

func hasTagManagePermission(c *fiber.Ctx) (bool, error) {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// Only authorized user is allowed to manage type
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "manage_tag"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return false, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to manage tag but he is not admin and has no \"manage_tag\" permission", stuId)
		return false, c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
	return true, nil
}

func GetTagList(c *fiber.Ctx) error {
	tagPrefix := c.Query("tagPrefix", "")
	offset, erro := strconv.Atoi(c.Query("offset", "0"))
	limit, errl := strconv.Atoi(c.Query("limit", "20"))
	if erro != nil || errl != nil || offset < 0 || limit < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "offset and limit must be integer"})
	}

	var tags []m.Tag
	if err := db.DB.Where("name LIKE ?", tagPrefix+"%").Offset(offset).Limit(limit).Find(&tags).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Fetched tag list",
		"data":    tags,
	})
}

func GetTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	hasBook := c.Query("hasBook", "false")
	var tag m.Tag
	if hasBook == "true" {
		if err := db.DB.Preload("Books").First(&tag, tagID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}
	} else {
		if err := db.DB.First(&tag, tagID).Error; err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Fetched tag",
		"data":    tag,
	})
}

func CreateTag(c *fiber.Ctx) error {
	var tag m.Tag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	if res, err := hasTagManagePermission(c); !res {
		return err
	}

	if err := db.DB.Create(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Created tag",
		"data":    tag,
	})
}

func UpdateTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	var tag m.Tag
	if err := c.BodyParser(&tag); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"message": "Invalid request body"})
	}

	if res, err := hasTagManagePermission(c); !res {
		return err
	}

	if err := db.DB.Model(&tag).Where("id = ?", tagID).Updates(tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Updated tag",
		"data":    tag,
	})
}

func DeleteTag(c *fiber.Ctx) error {
	tagID := c.Params("id")
	var tag m.Tag
	if err := db.DB.First(&tag, tagID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	if res, err := hasTagManagePermission(c); !res {
		return err
	}

	if err := db.DB.Delete(&tag).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"message": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Deleted tag",
		"data":    nil,
	})
}
