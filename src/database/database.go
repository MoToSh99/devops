package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"
)

//InitDB initialize the database tables
func InitDB(db *sql.DB) {

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

//ConnectDB returns a new connection to the database
func ConnectDB() *sql.DB {
	connection, err := sql.Open("sqlite3", "/tmp/minitwit.db")
	if err != nil {
		fmt.Println(err)
	}
	return connection
}

//ExecCommand executes sql command
func ExecCommand(sqlCommand string, db *sql.DB) {
	statement, err := db.Prepare(sqlCommand)
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()
}

//QueryDB queries the database and returns a list of rows
func QueryDB(query string, args []interface{}, db *sql.DB) *sql.Rows {
	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
	}
	return rows

}
