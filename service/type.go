package service

import (
	db "BooPT/database"
	m "BooPT/model"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strconv"
)

func hasTypeManagePermission(c *fiber.Ctx) (bool, error) {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := int64(jwtClaims["stu_id"].(float64))
	userId := uint(jwtClaims["sub"].(float64))
	userState := int(jwtClaims["state"].(float64))

	// Only authorized user is allowed to manage type
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "manage_type"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return false, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to manage type but he is not admin and has no \"manage_type\" permission", stuId)
		return false, c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Forbidden"})
	}
	return true, nil
}

func GetAllType(c *fiber.Ctx) error {
	offset, erro := strconv.Atoi(c.Query("offset", "0"))
	limit, errl := strconv.Atoi(c.Query("limit", "30"))
	if erro != nil || errl != nil || offset < 0 || limit < 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid query parameter"})
	}

	var types []m.Type
	if err := db.DB.Find(&types).Offset(offset).Limit(limit).Error; err != nil {
		logrus.Errorf("Error when get type list: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Get type list success",
		"data":    types,
	})
}

func GetType(c *fiber.Ctx) error {
	id := c.Params("id")
	var type_ m.Type
	if err := db.DB.First(&type_, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Type not found"})
		}
		logrus.Errorf("Error when get type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Get type success",
		"data":    type_,
	})
}

func CreateType(c *fiber.Ctx) error {
	var type_ m.Type
	if err := c.BodyParser(&type_); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if res, err := hasTypeManagePermission(c); !res {
		return err
	}

	if err := db.DB.Create(&type_).Error; err != nil {
		logrus.Errorf("Error when create type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Create type success",
		"data":    type_,
	})
}

func UpdateType(c *fiber.Ctx) error {
	var type_ m.Type
	if err := c.BodyParser(&type_); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if res, err := hasTypeManagePermission(c); !res {
		return err
	}

	var dbType m.Type
	if err := db.DB.First(&dbType, type_.ID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Type not found"})
		}
		logrus.Errorf("Error when get type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if type_.Name != "" {
		dbType.Name = type_.Name
	}
	dbType.Description = type_.Description

	if err := db.DB.Save(&dbType).Error; err != nil {
		logrus.Errorf("Error when update type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Update type success",
		"data":    type_,
	})
}

func DeleteType(c *fiber.Ctx) error {
	id := c.Params("id")
	var type_ m.Type
	if err := db.DB.First(&type_, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Type not found"})
		}
		logrus.Errorf("Error when get type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	if res, err := hasTypeManagePermission(c); !res {
		return err
	}

	if err := db.DB.Delete(&type_).Error; err != nil {
		logrus.Errorf("Error when delete type: %#v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Delete type success",
		"data":    nil,
	})
}
