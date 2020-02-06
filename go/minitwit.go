package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

var DATABASE = connect_db()
var SECRET_KEY = []byte("development key")
var STORE = sessions.NewCookieStore(SECRET_KEY)

func main() {

	fmt.Println("Hello Mack")

	r := mux.NewRouter()
	r.HandleFunc("/", timeline)
	r.HandleFunc("/public", public_timeline)
	r.HandleFunc("/{username}", user_timeline)
	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/logout", logout)

	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")

	http.ListenAndServe(":5000", r)

}

//Returns a new connection to the database
func connect_db() *sql.DB {
	connection, err := sql.Open("sqlite3", "/tmp/minitwit.db")
	if err != nil {
		fmt.Println(err)
	}
	return connection
}

func init_db() {
	//Initialize the database tables
	file, err := os.Open("schema.sql")
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
		if currentLine == ";" {
			execCommand(statement)
			statement = "" //Reset statement string
		}
	}
}

func execCommand(sqlCommand string) {
	statement, err := DATABASE.Prepare(sqlCommand)
	if err != nil {
		fmt.Println(err)
	}
	statement.Exec()
}

func query_db(query string, args []string, one bool) {
	//Query the database and returns a list of dictionaries
}

func get_user_id(username string) int {
	//Convenience method to loop up the id for a username
	var id int
	rows, err := DATABASE.Query("select user_id from user where username = ?", username)
	if err != nil {
		fmt.Println(err)
	}
	for rows.Next() {
		err := rows.Scan(&id)

		if err != nil {
			fmt.Println(err)
		}
	}
	return id
}

func format_datetime(timestamp int) {
	//Format a timestamp for display
}

func gravatar_url(email string, size int) {
	//Return the gravatar image for the given email address
}

func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Timeline hit")
}

func public_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "public_timeline hit")
}

func user_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "user_timeline hit")
}
func follow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "follow_user hit")
}
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "unfollow_user")
}

func add_message(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "add_message hit")
}

func login(w http.ResponseWriter, r *http.Request) {
	session, err := STORE.Get(r, "session")

	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, "login hit")
	username := r.FormValue("username")
	//password := r.FormValue("password")
	//redirectTarget := "timeline"
	user, err := DATABASE.Query("select * from user where username = ?", username)
	if user.Err != nil { //Jeg skal tjekke om jeg har fundet en bruger. Idk man
		fmt.Println("Invalid username")
	}
	//check password hash osv osv.

	session.AddFlash("You were logged in")

}

func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "logout hit")
	session, err := STORE.Get(r, "session")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash("You were logged out")
	session.Values["user_id"] = nil
	sessions.Save(r, w)
	http.Redirect(w, r, "public_timeline", http.StatusFound)

}

func register(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "register hit")
}
