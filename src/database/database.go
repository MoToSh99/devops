package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

func Init_db(db *sql.DB) {
	//Initialize the database tables
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
			ExecCommand(statement, db)
			statement = "" //Reset statement string
		}
	}
}

//Returns a new connection to the database
func Connect_db() *sql.DB {
	connection, err := sql.Open("sqlite3", "/tmp/minitwit.db")
	if err != nil {
		fmt.Println(err)
	}
	return connection
}

func ExecCommand(sqlCommand string, db *sql.DB) {
	statement, err := db.Prepare(sqlCommand)
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()
}

func Query_db(query string, args []string, one bool) {
	//Query the database and returns a list of dictionaries
}
