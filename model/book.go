package model

import "gorm.io/gorm"

type Book struct {
	gorm.Model
	BID           int64          `json:"book_id" gorm:"uniqueIndex;column:bid"` // positive: ISBN (10 digits) || negative: internal id for school based books
	Title         string         `json:"title"`                                 // book name
	Author        string         `json:"author"`                                // book author
	Publisher     string         `json:"publisher" gorm:"index"`                // book publish company
	Version       string         `json:"version"`                               // book version
	State         BookState      `json:"state"`                                 // book state
	Image         string         `json:"image"`                                 // book image
	Description   string         `json:"description"`                           // book description
	TypeId        int64          `json:"type_id" gorm:"index"`                  // book type id
	Type          Type           `json:"type"`                                  // book type
	DownloadLinks []DownloadLink `json:"download_links" gorm:"references:BID"`  // book download links
	Tags          []*Tag         `json:"tags" gorm:"many2many:book_tags"`       // book tags
}

type BookState int64

const (
	BookState_Unpublished BookState = iota
	BookState_Published             // 已发布
	BookState_Undeveloped           // 暂时下架
	BookState_Deleted               // 已删除
)

type DownloadLink struct {
	gorm.Model
	BookId          int64  `json:"book_bid" gorm:"index"` // book bid
	Link            string `json:"link"`
	VersionDescribe string `json:"version_describe"`
	Password        string `json:"password"`
	State           int64  `json:"state"`
}

type Tag struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Books       []*Book `json:"books" gorm:"many2many:book_tags"`
	Description string  `json:"description"`
}

type Type struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
