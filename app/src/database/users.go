package database

import (
	"github.com/matt035343/devops/app/src/types"
)

//AddUser Adds a new user entry to the database.
func (d *Database) AddUser(username, email, hash string) error {
	return d.db.Create(&types.User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
	}).Error
}

//GetUser Queries user information in the database given userID.
func (d *Database) GetUser(userID int) (user types.User, err error) {
	err = d.db.Where(&types.User{UserID: userID}).First(&user).Error
	return user, err
}

//GetUserFromUsername Queries user information in the database given username.
func (d *Database) GetUserFromUsername(username string) (user types.User, err error) {
	err = d.db.Where(&types.User{Username: username}).First(&user).Error
	return user, err
}
