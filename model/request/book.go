package request

type Book struct {
	Title         string  `json:"title"`
	Author        string  `json:"author"`
	Tags          []int64 `json:"tags"`
	TypeId        int64   `json:"type_id" gorm:"index"` // book type id
	BID           int64   `json:"bid"`
	Publisher     string  `json:"publisher" gorm:"index"` // book publish company
	Version       string  `json:"version"`                // book version
	State         int64   `json:"state"`                  // book state
	Image         string  `json:"image"`                  // book image
	Description   string  `json:"description"`            // book description
	PublishReason string  `json:"publish_reason"`         // book publish reason
	Offset        int     `json:"offset"`
	Limit         int     `json:"limit"`
}
