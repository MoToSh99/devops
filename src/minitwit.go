package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	"go/src/database"
	"go/src/utils"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DATABASE = database.Connect_db()
var SECRET_KEY = []byte("development key")
var STORE = sessions.NewCookieStore(SECRET_KEY)
var PER_PAGE = 30
var STATIC_ROOT_PATH = "./src/static"

func main() {

	// Init DB if it doesn't exist
	if !utils.FileExists("/tmp/minitwit.db") {
		fmt.Println("Initializing database")
		database.Init_db(DATABASE)
	}
	fmt.Println("Running: localhost:5000/public")
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
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
	// http.Handle("/", Before_request(r))
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
	IsLoggedIn      bool
	SessionUser     string
	UserProfile     string
	Followed        bool
}

type User struct {
	User_id  int
	Username string
	Email    string
	Pw_hash  string
}

func Before_request(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("middleware", r.URL)
		getSession(w, r)
	})
}

func getSession(w http.ResponseWriter, r *http.Request) *sessions.Session {
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return session
}

func get_user_id(username string) int {
	//Convenience method to loop up the id for a username
	var id int
	var rows *sql.Rows = database.Query_db("select user_id from user where username = ?", []string{username}, DATABASE)
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
		http.Redirect(w, r, "/public", http.StatusNotFound)
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
		http.Redirect(w, r, "/public_timeline", http.StatusForbidden)
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
		IsLoggedIn:      false,
		SessionUser:     (user.(*User)).Username,
		UserProfile:     "",
	}

	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/timeline.html"))

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

	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/timeline.html"))

	tmpl.Execute(w, data)
}

func user_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("User timeline hit")
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := session.Values["user"]
	username := (user.(*User)).Username
	user_id := (user.(*User)).User_id

	stmt, err := DATABASE.Prepare("SELECT * FROM user WHERE username = ?")
	profile_user_id := -999
	profile_email := ""
	profile_pw_hash := ""
	profile_username := ""
	err = stmt.QueryRow(username).Scan(&profile_user_id, &profile_username, &profile_email, &profile_pw_hash)
	if err != nil && err.Error() != "sql: no rows in result set" {
		panic(err)
	}

	stmt, err = DATABASE.Prepare(`select 1 from follower where
	follower.who_id = ? and follower.whom_id = ?`)
	var who_id int
	var whom_id int
	var followed bool
	err = stmt.QueryRow(user_id, profile_user_id).Scan(&who_id, &whom_id)
	if err != nil && err.Error() != "sql: no rows in result set" {
		followed = false
	}
	followed = true

	var rows *sql.Rows = database.Query_db(`select message.*, user.* from message, user where
	user.user_id = message.author_id and user.user_id = ?
	order by message.pub_date desc limit ?`, []string{string(profile_user_id), string(PER_PAGE)}, DATABASE)
	fmt.Println("query run")
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
			fmt.Println(err)
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
		RequestEndpoint: "user_timeline",
		Messages:        messages,
		IsLoggedIn:      true,
		SessionUser:     username,
		UserProfile:     profile_username,
		Followed:        followed,
	}

	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/timeline.html"))

	tmpl.Execute(w, data)

}
func follow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "follow_user hit")
}
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "unfollow_user")
}

func add_message(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "add_message hit")
	session, err := STORE.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user := session.Values["user"]
	if user == nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	text := r.FormValue("text")
	user_id := (user.(*User)).User_id

	if text != "" {
		queryString := `INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`
		statement, err := DATABASE.Prepare(queryString)
		_, err = statement.Exec(user_id, text, time.Now())
		checkErr(err)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		loginGet(w, r)
	} else if r.Method == "POST" {
		loginPost(w, r)
	}

}

func loginGet(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/login.html"))
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
		HasError        bool
		ErrorMsg        string
		IsLoggedIn      bool
		Username        string
		RequestEndpoint string
	}{false, errorMsg, false, "", ""}

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
		tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/login.html"))
		data.HasError = true
		data.ErrorMsg = "Cannot recognize user"
		data.IsLoggedIn = false
		tmpl.Execute(w, data)
		return
	}

	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/timeline.html"))
	data.IsLoggedIn = true
	data.Username = user.Username
	data.RequestEndpoint = "public_timeline"
	tmpl.Execute(w, data)
	//http.Redirect(w, r, "/public_timeline", http.StatusFound)

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
	fmt.Println("logout hit")
	session, err := STORE.Get(r, "session")
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.AddFlash("You were logged out")

	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		panic(err)
	}
	sessions.Save(r, w)

	// tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/timeline.html"))
	// tmpl.Execute(w, nil)
	http.Redirect(w, r, "/public", http.StatusFound)

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
	tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/register.html"))
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

	if errorMsg != "" {
		tmpl := template.Must(template.ParseFiles(STATIC_ROOT_PATH + "/templates/register.html"))
		data := struct {
			HasError   bool
			ErrorMsg   string
			IsLoggedIn bool
		}{true, errorMsg, false}
		tmpl.Execute(w, data)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

}
