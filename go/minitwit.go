package main

import (
	"bufio"
	"database/sql"
	"encoding/gob"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DATABASE = connect_db()
var SECRET_KEY = []byte("development key")
var STORE = sessions.NewCookieStore(SECRET_KEY)
var PER_PAGE = 30

func main() {

	fmt.Println("Running: localhost:5000/public")
	//init_db()
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("static/css/"))),
	)
	gob.Register(&User{})

	r.HandleFunc("/", timeline)
	r.HandleFunc("/public", public_timeline)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")
	r.HandleFunc("/{username}", user_timeline)
	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
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

type User struct {
	User_id  int
	Username string
	Email    string
	Pw_hash  string
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
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := session.Values["user"]
	if user == nil {
		http.Redirect(w, r, "/public", http.StatusFound)
	}
	user_id := (user.(*User)).User_id
	stmt, err := DATABASE.Prepare(`select message.*, user.* from message, user
	where message.flagged = 0 and message.author_id = user.user_id and (
		user.user_id = ? or
		user.user_id in (select whom_id from follower
								where who_id = ?))
	order by message.pub_date desc limit ?`)
	rows, err := stmt.Query(user_id, user_id, PER_PAGE)
	if err != nil {
		panic(err)
	}

	messages := []MessageViewData{}

	for rows.Next() {
		var message_id int
		var author_id int
		var text string
		var pub_date int64
		var flagged int

		var user_id int
		var username string
		var email string
		var pw_hash string

		err = rows.Scan(&message_id, &author_id, &text, &pub_date, &flagged, &user_id, &username, &email, &pw_hash)
		if err != nil {
			log.Fatal(err)
		}

		message := MessageViewData{
			Text:         text,
			Email:        email,
			Gravatar_url: gravatar_url(email, 64),
			Username:     username,
			Pub_date:     time.Unix(pub_date, 0).String(),
		}
		messages = append(messages, message)
	}

	data := RequestData{
		Title:           "title",
		RequestEndpoint: "public_timeline",
		Messages:        messages,
	}

	tmpl := template.Must(template.ParseFiles("./static/templates/timeline.html"))

	tmpl.Execute(w, data)

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
	if r.Method == "GET" {
		loginGet(w, r)
	} else if r.Method == "POST" {
		loginPost(w, r)
	}

}

func loginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/templates/login.html"))
	tmpl.Execute(w, nil)
}

func loginPost(w http.ResponseWriter, r *http.Request) {
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	errorMsg := ""
	data := struct {
		HasError bool
		ErrorMsg string
	}{false, errorMsg}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" {
		errorMsg = "You have to enter a username"
	}
	if password == "" {
		errorMsg = "You have to enter a password"
	}

	userFound, user := Authenticate(username, password)
	if !userFound {
		tmpl := template.Must(template.ParseFiles("./static/templates/login.html"))
		data.HasError = true
		data.ErrorMsg = "Cannot recognize user"
		tmpl.Execute(w, data)
		return
	}

	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)

}

func Authenticate(username string, password string) (bool, *User) {

	stmt, err := DATABASE.Prepare("SELECT * FROM user WHERE username = ?")
	user_id := -999
	email := ""
	pw_hash := ""
	err = stmt.QueryRow(username).Scan(&user_id, &username, &email, &pw_hash)

	user := &User{
		User_id:  user_id,
		Username: username,
		Email:    email,
		Pw_hash:  pw_hash,
	}

	if err != nil && err.Error() != "sql: no rows in result set" {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(pw_hash), []byte(password))
	if user_id == -999 || pw_hash == "" || err != nil {
		return false, nil
	}
	return true, user

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
	http.Redirect(w, r, "/", http.StatusFound)

}

func register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		registerGet(w, r)
	} else if r.Method == "POST" {
		registerPost(w, r)
	}

}

func isUsernameTaken(username string) bool {
	return false
}

func registerGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("./static/templates/register.html"))
	tmpl.Execute(w, nil)
}

func registerUser(username string, email string, hashedPassword string) bool {
	queryString := "INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)"
	statement, err := DATABASE.Prepare(queryString)
	_, err = statement.Exec(username, email, hashedPassword)
	checkErr(err)

	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func registerPost(w http.ResponseWriter, r *http.Request) {
	errorMsg := ""

	if r.FormValue("username") == "" {
		errorMsg = "You have to enter a username"
	} else if r.FormValue("email") == "" || !strings.Contains(r.FormValue("email"), "@") {
		errorMsg = "You have to enter a valid email address"
	} else if r.FormValue("password") == "" {
		errorMsg = "You have to enter a password"
	} else if r.FormValue("password") != r.FormValue("password2") {
		errorMsg = "The two passwords do not match"
	} else if isUsernameTaken(r.FormValue("username")) {
		errorMsg = "The username is already taken"
	} else {
		hashedPasswordInBytes, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 14)
		registerUser(r.FormValue("username"), r.FormValue("email"), string(hashedPasswordInBytes))
	}

	tmpl := template.Must(template.ParseFiles("./static/templates/register.html"))
	data := struct {
		HasError bool
		ErrorMsg string
	}{true, errorMsg}
	tmpl.Execute(w, data)
}
