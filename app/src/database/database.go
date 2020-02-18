package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/matt035343/devops/src/types"
	"github.com/matt035343/devops/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

var db = ConnectDB()

//InitDB initialize the database tables
func InitDB() {
	file, err := os.Open("src/database/schema.sql")
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

//ConnectDB returns a new connection to the database
func ConnectDB() *sql.DB {
	connection, err := sql.Open("sqlite3", "/tmp/minitwit.db")
	if err != nil {
		fmt.Println(err)
	}
	return connection
}

//ExecCommand executes sql command
func ExecCommand(sqlCommand string) {
	statement, err := db.Prepare(sqlCommand)
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()
}

//QueryDB queries the database and returns a list of rows
func QueryRowDB(query string, args ...interface{}) *sql.Row {
	return db.QueryRow(query, args...) //automatically preparing
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
