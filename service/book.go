package service

import (
	db "BooPT/database"
	m "BooPT/model"
	r "BooPT/model/request"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func GetBook(c *fiber.Ctx) error {
	bookID := c.Params("id")
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := jwtClaims["stu_id"].(int64)
	userState := jwtClaims["state"].(int)
	var book m.Book
	// Get book from database
	if err := db.DB.Joins("Tags").Joins("Types").First(&book, bookID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("User %v want to get book %v but it doesn't exist", c.Get("userID"), bookID)
			return c.Status(404).JSON(fiber.Map{"error": "Book not found"})
		} else {
			logrus.Error(err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}
	}
	// If book is not public and user is not owner of book
	if book.State != m.BookState_Published {
		var publishRecord m.PublishRecord
		// Get publish record from database
		if err := db.DB.Where("book_id = ?", bookID).First(&publishRecord).Error; err != nil {
			logrus.Errorf("Book %v should has publish_record but it doesn't exist", bookID)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}
		// If user is not owner of book or is admin
		if publishRecord.UserId != stuId && userState <= 0 {
			logrus.Warnf("User %v want to get book %v but it is not public", stuId, bookID)
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}
	}
	// Return book
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Book found",
		"data":    book,
	})
}

func GetBookList(c *fiber.Ctx) error {
	var books []m.Book
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := jwtClaims["stu_id"].(int64)
	userId := jwtClaims["sub"].(uint)
	userState := jwtClaims["state"].(int)

	// Parse request
	var bookReq r.Book
	if err := c.QueryParser(&bookReq); err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if bookReq.Limit == 0 {
		bookReq.Limit = 20
	}

	// Start query book id match condition
	var tx *gorm.DB
	if c.Query("review_mode", "false") == "true" {
		var user m.User
		if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "review_book"})).First(&user, userId).Error; err != nil {
			logrus.Error(err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}
		if !(userState > 0 || len(user.Permissions) > 0) {
			logrus.Warnf("User %v want to get unpub book list %v but he is not admin and has no \"review_book\" permission", stuId, bookReq.Title)
			return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
		}
		tx = db.DB.Where(m.Book{State: m.BookState(bookReq.State)})
	} else {
		tx = db.DB.Where(m.Book{State: m.BookState_Published})
	}

	if bookReq.TypeId != 0 { // limit by type
		tx = tx.Joins("Types", "id = ?", bookReq.TypeId)
	} else {
		tx = tx.Joins("Types")
	}
	if len(bookReq.Tags) > 0 { // limit by tags, all tags must be in book
		tx = tx.Preload("Tags", db.DB.Where("id IN (?)", bookReq.Tags)).Group("id").Having("COUNT(id) = ?", len(bookReq.Tags))
	} else {
		tx = tx.Preload("Tags")
	}
	if len(bookReq.Title) > 0 { // limit by title
		tx = tx.Where("title LIKE ?", "%"+bookReq.Title+"%")
	}
	if len(bookReq.Author) > 0 { // limit by author
		tx = tx.Where("author LIKE ?", "%"+bookReq.Author+"%")
	}
	if len(bookReq.Publisher) > 0 { // limit by publisher
		tx = tx.Where("publisher LIKE ?", "%"+bookReq.Publisher+"%")
	}
	if bookReq.BID != 0 { // limit by book id
		tx = tx.Where("bid = ?", bookReq.BID)
	}
	if err := tx.Select("id").Offset(bookReq.Offset).Limit(bookReq.Limit).Find(&books).Error; err != nil {
		logrus.Errorf("User %v want to get book list but error happend, %#v ", stuId, err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	} else if len(books) == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Books not found"})
	}

	// Get book info from database
	idList := make([]uint, len(books))
	for i, book := range books {
		idList[i] = book.ID
	}
	if err := db.DB.Find(&books, idList).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("User %v want to get book list but there is no book", stuId)
			return c.Status(404).JSON(fiber.Map{"error": "No Book found"})
		} else {
			logrus.Error(err)
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Book list found",
		"data":    books,
	})
}

func UploadNewBook(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := jwtClaims["stu_id"].(int64)
	userId := jwtClaims["sub"].(uint)
	userState := jwtClaims["state"].(int)
	// parse request body
	var bookReq r.Book
	if err := c.BodyParser(&bookReq); err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Check if user is admin or has upload permission, warning: hard code 'upload_book' permission
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "upload_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to upload book but he is not admin and has no \"upload_book\" permission", stuId)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// get tags for book
	var tags []m.Tag
	if err := db.DB.Find(&tags, bookReq.Tags).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if len(tags) != len(bookReq.Tags) {
		logrus.Warnf("User %v want to upload book but some tags are not exist", stuId)
		return c.Status(400).JSON(fiber.Map{"error": "Some tags are not exist"})
	}
	pTags := make([]*m.Tag, len(tags))
	for i, tag := range tags {
		pTags[i] = &tag
	}

	// create book
	book := m.Book{
		BID:         bookReq.BID,
		Title:       bookReq.Title,
		Author:      bookReq.Author,
		Publisher:   bookReq.Publisher,
		Version:     bookReq.Version,
		State:       m.BookState_Unpublished,
		Image:       bookReq.Image,
		Description: bookReq.Description,
		TypeId:      bookReq.TypeId,
		Tags:        pTags,
	}

	// Check if book is already exist
	if err := db.DB.Where("bid = ?", book.BID).First(&book).Error; err == nil {
		logrus.Warnf("User %v want to upload book %v but it already exist", stuId, book.Title)
		return c.Status(409).JSON(fiber.Map{"error": "Book ISBN already exist"})
	}

	// Save book to db and create publish record
	tx := db.DB.Begin()
	if err := tx.Create(&book).Error; err != nil {
		logrus.Errorf("User %v want to create book %v but it failed, %#v ", stuId, book.Title, err)
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	publishRecord := m.PublishRecord{
		BookId:        book.ID,
		UserId:        stuId,
		PublishReason: bookReq.PublishReason,
	}
	if err := tx.Create(&publishRecord).Error; err != nil {
		logrus.Errorf("Create publish record for user %v book %v but failed, %#v ", stuId, book.Title, err)
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if err := tx.Commit().Error; err != nil {
		logrus.Errorf("Commit user %v book %v failed due to %#v", stuId, book.Title, err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Book created",
		"data":    book,
	})
}

func UpdateBook(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := jwtClaims["stu_id"].(int64)
	userId := jwtClaims["sub"].(uint)
	userState := jwtClaims["state"].(int)
	// parse request body
	var bookReq r.Book
	if err := c.BodyParser(&bookReq); err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Check if user is admin or has update permission, warning: hard code 'review_book' permission
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "review_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to update book %v but he is not admin and has no \"review_book\" permission", stuId, bookReq.Title)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get book
	var book m.Book
	if err := db.DB.Where("bid = ?", bookReq.BID).First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("User %v want to update book %v but it does not exist", stuId, bookReq.BID)
			return c.Status(404).JSON(fiber.Map{"error": "Book not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Get Tags
	var tags []m.Tag
	if err := db.DB.Find(&tags, bookReq.Tags).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	pTags := make([]*m.Tag, len(tags))
	for i, tag := range tags {
		pTags[i] = &tag
	}

	// Modify book, maybe there is a better way to do this...
	book.Title = m.UpdateIfNotEmpty(bookReq.Title, book.Title).(string)
	book.Author = m.UpdateIfNotEmpty(bookReq.Author, book.Author).(string)
	book.TypeId = m.UpdateIfNotEmpty(bookReq.TypeId, book.TypeId).(int64)
	book.BID = m.UpdateIfNotEmpty(bookReq.BID, book.BID).(int64)
	book.Publisher = m.UpdateIfNotEmpty(bookReq.Publisher, book.Publisher).(string)
	book.Version = m.UpdateIfNotEmpty(bookReq.Version, book.Version).(string)
	book.Image = m.UpdateIfNotEmpty(bookReq.Image, book.Image).(string)
	book.Description = m.UpdateIfNotEmpty(bookReq.Description, book.Description).(string)
	book.State = m.BookState(bookReq.State)

	// Save book to db
	if err := db.DB.Save(&book).Error; err != nil {
		logrus.Errorf("User %v want to update book %v but it failed, %#v ", stuId, book.Title, err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if len(pTags) > 0 {
		// Replace tags
		if err := db.DB.Model(&book).Association("Tags").Replace(pTags).Error; err != nil {
			logrus.Errorf("User %v want to update book %v but it failed, %#v", stuId, book.Title, err())
			return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
		}
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Book updated",
		"data":    book,
	})
}

func DeleteBook(c *fiber.Ctx) error {
	jwtClaims := c.Locals("jwt").(*jwt.Token).Claims.(jwt.MapClaims)
	stuId := jwtClaims["stu_id"].(int64)
	userId := jwtClaims["sub"].(uint)
	userState := jwtClaims["state"].(int)
	// parse request body
	var bookReq r.Book
	if err := c.QueryParser(&bookReq); err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Check if user is admin or has "delete permission", warning: hard code 'delete_book' permission
	var user m.User
	if err := db.DB.Preload("Permissions", db.DB.Where(&m.Permission{Name: "delete_book"})).First(&user, userId).Error; err != nil {
		logrus.Error(err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}
	if !(userState > 0 || len(user.Permissions) > 0) {
		logrus.Warnf("User %v want to delete book %v but he is not admin and has no \"delete_book\" permission", stuId, bookReq.Title)
		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}

	// Get book
	var book m.Book
	if err := db.DB.Where("bid = ?", bookReq.BID).First(&book).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logrus.Warnf("User %v want to delete book %v but it does not exist", stuId, bookReq.BID)
			return c.Status(404).JSON(fiber.Map{"error": "Book not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	// Delete book
	if err := db.DB.Delete(&book).Error; err != nil {
		logrus.Errorf("User %v want to delete book %v but it failed, %#v ", stuId, book.Title, err)
		return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Book deleted",
		"data":    nil,
	})
}
