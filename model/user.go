package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email           string           `gorm:"type:varchar(63);uniqueIndex"`
	Name            string           `gorm:"type:varchar(31)"`
	StuId           int64            `json:"student_id" gorm:"uniqueIndex"`
	State           int              `gorm:"default:0"` // -1: locked, 0: normal, 1: admin
	Permissions     []*Permission    `gorm:"many2many:user_permissions"`
	DownloadLimit   int64            `gorm:"default:20"`
	DownloadRecords []DownloadRecord `gorm:"foreignkey:UserId;references:StuId"`
	RequestRecords  []RequestRecord  `gorm:"foreignkey:UserId;references:StuId"`
	PasswordHash    string           `json:"password_hash"`
}

type Permission struct {
	gorm.Model
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Users       []*User `gorm:"many2many:user_permissions"`
}
