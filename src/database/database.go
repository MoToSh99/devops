package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"go/src/types"
	"go/src/utils"
	"log"
	"os"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

//InitDB initialize the database tables
func initDatabase(schemaLocation string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	databasePath := wd + schemaLocation
	file, err := os.Open(databasePath)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentLine string
	statement := ""
	for scanner.Scan() { //Read lines in schema.sql until semicolon which triggers the execute command.
		currentLine = scanner.Text()
		statement = statement + currentLine
		if strings.Contains(currentLine, ";") {
			ExecCommand(statement)
			statement = "" //Reset statement string
		}
	}
}

func ConnectDatabase(databaseDialect, connectionString string, createDatabaseIfAbsent bool, schemaLocation string) (*gorm.DB, err) {
	if databaseDialect == "sqlite3" && createDatabaseIfAbsent&!utils.FileExists(connectionString) {
		initDatabase(schemaLocation)
	}
	return gorm.Open(databaseDialect, connectionString)
}

//QueryDB queries the database and returns a list of rows
func (db *gorm.DB) getFollowers(userID int) ([]types.Follower, error) {
	var followers []type.Follower
	err := db.Where(&types.Follower{WhomID: userID}).Find(&followers).Error
	return followers, err
}

//QueryDB queries the database and returns a list of rows
func QueryRowsDB(query string, args ...interface{}) *sql.Rows {
	rows, err := db.Query(query, args...) //automatically preparing
	if err != nil {
		fmt.Println(err)
	}
	return rows
}

func AlterDB(query string, args ...interface{}) error {
	statement, err := db.Prepare(query)
	_, err = statement.Exec(args...)
	return err
}

func QueryMessages(query string, args ...interface{}) []types.MessageViewData {
	rows := QueryRowsDB(query, args...)

	messages := []types.MessageViewData{}

	for rows.Next() {
		message := types.Message{}
		messageUser := types.User{}

		err := rows.Scan(&message.MessageID, &message.AuthorID, &message.Text, &message.PublishedDate, &message.Flagged, &messageUser.UserID, &messageUser.Username, &messageUser.Email, &messageUser.PasswordHash)
		if err != nil {
			log.Fatal(err)
		}

		messageViewData := types.MessageViewData{
			Text:          message.Text,
			Email:         messageUser.Email,
			GravatarURL:   utils.GravatarURL(messageUser.Email, 48),
			Username:      messageUser.Username,
			PublishedDate: utils.FormatDatetime(message.PublishedDate),
		}
		messages = append(messages, messageViewData)
	}
	return messages
}

func QueryFollowers(query string, args ...interface{}) types.FollowerResponse {
	rows := QueryRowsDB(query, args...)
	followers := []string{}

	for rows.Next() {
		follower := ""
		err := rows.Scan(&follower)
		if err != nil {
			log.Fatal(err)
		}
		followers = append(followers, follower)
	}
	return types.FollowerResponse{Follows: followers}
}
