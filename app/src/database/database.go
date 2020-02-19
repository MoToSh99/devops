package database

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/matt035343/devops/src/types"
	"github.com/matt035343/devops/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *gorm.DB
}

func New(gdb *gorm.DB) *Database {
	return &Database{db: gdb}
}

func ConnectDatabase(databaseDialect, connectionString string) (*Database, error) {
	db, err := gorm.Open(databaseDialect, connectionString)
	autoMigrate(db)
	return New(db), err
}

func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&types.User{}).AutoMigrate(&types.Message{}).AutoMigrate(&types.Follower{}).Error
}

func (d *Database) GetFollower(whoID, whomID int) (follower types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: whomID, WhoID: whoID}).First(&follower).Error
	return follower, err
}

func (d *Database) GetFollowers(userID, limit int) (followers []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhomID: userID}).Limit(limit).Find(&followers).Error
	return followers, err
}

func (d *Database) GetFollowing(userID, limit int) (following []types.Follower, err error) {
	err = d.db.Where(&types.Follower{WhoID: userID}).Limit(limit).Find(&following).Error
	return following, err
}

func (d *Database) GetUser(userID int) (user types.User, err error) {
	err = d.db.Where(&types.User{UserID: userID}).First(&user).Error
	return user, err
}

func (d *Database) GetUserFromUsername(username string) (user types.User, err error) {
	err = d.db.Where(&types.User{Username: username}).First(&user).Error
	return user, err
}

func (d *Database) GetMessages(userID, limit int) (messages []types.Message, err error) {
	err = d.db.Where(&types.Message{AuthorID: userID}).Limit(limit).Find(&messages).Error
	return messages, err
}

func (d *Database) GetPublicViewMessages(limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Where("flagged = ", false).Limit(limit).Order("published_date desc").Find(&ms).Error
	if err != nil {
		return messages, err
	}
	messages = d.convertMessageModelsToViewModels(ms)
	return messages, nil
}

func (d *Database) GetUserViewMessages(userID, limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Where(&types.Message{AuthorID: userID}).Where("flagged = ", false).Limit(limit).Order("published_date desc").Find(&ms).Error
	if err != nil {
		return messages, err
	}
	messages = d.convertMessageModelsToViewModels(ms)
	return messages, nil
}

func (d *Database) GetTimelineViewMessages(userID, limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Where("user.user_id = ? or user.user_id in (select whom_id from followers where who_id = ?)", userID, userID).Limit(limit).Order("published_date desc").Find(&ms).Error
	if err != nil {
		return messages, err
	}
	messages = d.convertMessageModelsToViewModels(ms)
	return messages, nil
}

func (d *Database) convertMessageModelsToViewModels(ms []types.Message) (messages []types.MessageViewData) {
	for _, m := range ms {
		user, err := d.GetUser(m.AuthorID)
		if err != nil {
			continue
		}
		message := types.MessageViewData{
			Text:          m.Text,
			Email:         user.Email,
			GravatarURL:   utils.GravatarURL(user.Email, 48),
			Username:      user.Username,
			PublishedDate: time.Unix(m.PublishedDate, 0).Format("dd-mm-YYYY"),
		}
		messages = append(messages, message)
	}
	return messages
}

func (d *Database) AddFollower(whoID, whomID int) error {
	f, _ := d.GetFollower(whoID, whomID)
	if !f.IsValidRelation() {
		return d.db.Create(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
	}
	return fmt.Errorf("Already following this user")
}

func (d *Database) DeleteFollower(whoID, whomID int) error {
	return d.db.Delete(&types.Follower{WhoID: whoID, WhomID: whomID}).Error
}

func (d *Database) AddMessage(authorID int, message string, time time.Time) error {
	return d.db.Create(&types.Message{
		Text:          message,
		AuthorID:      authorID,
		PublishedDate: time.Unix(),
		Flagged:       false,
	}).Error
}

func (d *Database) AddUser(username, email, hash string) error {
	return d.db.Create(&types.User{
		Email:        email,
		Username:     username,
		PasswordHash: hash,
	}).Error
}
