package model

import (
	"gorm.io/gorm"
	"time"
)

type DownloadRecord struct {
	ID        uint64    `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UserId    int64     `gorm:"index"`
	User      User
	BookId    uint `gorm:"index"`
	Book      Book
}

type RequestRecord struct {
	gorm.Model
	UserId      int64 `gorm:"index"` // student id
	User        User
	BookName    string
	BID         int64 `gorm:"column:bid"` // ISBN number or others(0)
	Description string
	PictureURL  string
}

type PublishRecord struct {
	gorm.Model
	UserId        int64 `gorm:"index"` // student id
	User          User
	BookId        uint `gorm:"uniqueIndex"`
	Book          Book
	PublishReason string
}

type UploadRecord struct {
	gorm.Model
	UserId         int64 `gorm:"index"` // student id
	User           User
	BookId         uint `gorm:"index"`
	Book           Book
	DownloadLinkID uint `gorm:"index"`
	DownloadLink   DownloadLink
	UploadReason   string
}

type FavoriteRecord struct {
	gorm.Model
	UserId int64 `gorm:"index"` // student id
	User   User
	BookId uint `gorm:"index"`
	Book   Book
}
