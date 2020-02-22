package types

type Message struct {
	// Author        User `gorm:"foreignkey:AuthorID;association_foreignkey:UserID"`
	ID            int `gorm:"primary_key"`
	AuthorID      int
	Text          string
	Flagged       bool
	MessageID     int
	PublishedDate int64
}
