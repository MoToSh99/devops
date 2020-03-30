package database

import (
	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/types"
)

//AddUser Adds a new user entry to the database.
func (d *Database) AddUser(username, email, hash string) error {
	err := d.db.Create(&types.User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
	}).Error
	log.ErrorErr("Could not create user with username %s and email %s", err, username, email)
	return err
}

//GetUser Queries user information in the database given userID.
func (d *Database) GetUser(userID int) (user types.User, err error) {
	err = d.db.Where(&types.User{UserID: userID}).First(&user).Error
	log.ErrorErr("Could not get user with userID %d", err, userID)
	return user, err
}

//GetUserFromUsername Queries user information in the database given username.
func (d *Database) GetUserFromUsername(username string) (user types.User, err error) {
	err = d.db.Where(&types.User{Username: username}).First(&user).Error
	log.ErrorErr("Could not get user with username %s", err, username)
	return user, err
}

//IsUsernameAvailable Returns a boolean whether the given username already exists in the database.
func (d *Database) IsUsernameAvailable(username string) bool {
	_, err := d.GetUserFromUsername(username)
	return err != nil
}
