package types

type Message struct {
	// Author        User `gorm:"foreignkey:AuthorID;association_foreignkey:UserID"`
	AuthorID      int `gorm:"primary_key"`
	Text          string
	Flagged       bool
	MessageID     int
	PublishedDate int64
}
