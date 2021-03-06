package types

//User the data model to store information about a user.
type User struct {
	UserID       int `gorm:"primary_key"`
	Username     string
	Email        string
	PasswordHash string
	// Followers    []Follower `gorm:"foreignkey:WhomID;association_foreignkey:UserID"`
	// Following    []Follower `gorm:"foreignkey:WhoID;association_foreignkey:UserID"`
}
