package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"

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
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("static/css/"))),
	)

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

type MessageViewData struct {
	Text         string
	Email        string
	Gravatar_url string
	Username     string
	Pub_date     string
}

type RequestData struct {
	Title           string
	RequestEndpoint string
	Messages        []MessageViewData
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
		if strings.Contains(currentLine, ";") {
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

func gravatar_url(email string, size int) string {
	//Return the gravatar image for the given email address
	return email
}

func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Timeline hit")
}

func public_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: " + r.RemoteAddr)
	data := RequestData{
		Title:           "title",
		RequestEndpoint: "public_timeline",
		Messages: []MessageViewData{
			{Text: "tweet tweet", Email: "email", Gravatar_url: gravatar_url("https://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50", 64),
				Username: "ikke bent", Pub_date: "10/04/2190"},
			{Text: "tweet tweet", Email: "email", Gravatar_url: gravatar_url("https://www.gravatar.com/avatar/205e460b479e2e5b48aec07710c08d50", 64),
				Username: "bent", Pub_date: "10/04/2190"},
		},
	}

	tmpl := template.Must(template.ParseFiles("./static/templates/timeline.html"))

	tmpl.Execute(w, data)
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
