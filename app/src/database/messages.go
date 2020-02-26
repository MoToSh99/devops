package database

import (
	"time"

	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

func (d *Database) GetAllMessages() (messages []types.Message, err error) {
	err = d.db.Model(&types.Message{}).Find(&messages).Error
	return messages, err
}

func (d *Database) FlagMessage(messageID int) (err error) {
	var message types.Message
	err = d.db.Where(&types.Message{ID: messageID}).First(&message).Error
	if err != nil {
		return err
	}
	message.Flagged = true
	return d.db.Save(&message).Error
}

func (d *Database) GetMessages(userID, limit int) (messages []types.Message, err error) {
	err = d.db.Where(&types.Message{AuthorID: userID}).Limit(limit).Find(&messages).Error
	return messages, err
}

func (d *Database) GetPublicViewMessages(limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Where("flagged = ?", false).Limit(limit).Order("published_date desc").Find(&ms).Error
	if err != nil {
		return messages, err
	}
	messages = d.convertMessageModelsToViewModels(ms)
	return messages, nil
}

func (d *Database) GetUserViewMessages(userID, limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Where(&types.Message{AuthorID: userID}).Where("flagged = ?", false).Limit(limit).Order("published_date desc").Find(&ms).Error
	if err != nil {
		return messages, err
	}
	messages = d.convertMessageModelsToViewModels(ms)
	return messages, nil
}

func (d *Database) GetTimelineViewMessages(userID, limit int) (messages []types.MessageViewData, err error) {
	var ms []types.Message
	err = d.db.Table("messages").Where("messages.author_id = ? or messages.author_id in (select whom_id from followers where who_id = ?)", userID, userID).Limit(limit).Order("published_date desc").Find(&ms).Error
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
			PublishedDate: time.Unix(m.PublishedDate, 0).Format(time.RFC822),
		}
		messages = append(messages, message)
	}
	return messages
}

func (d *Database) AddMessage(authorID int, message string, time time.Time) error {
	return d.db.Create(&types.Message{
		Text:          message,
		AuthorID:      authorID,
		PublishedDate: time.Unix(),
		Flagged:       false,
	}).Error
}
