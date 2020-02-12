package main

import (
	"database/sql"
	"encoding/gob"
	"fmt"
	authentication "go/src/authentication"
	"go/src/database"
	"go/src/types"
	"go/src/utils"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var DATABASE = database.ConnectDB()
var PER_PAGE = 30
var STATIC_ROOT_PATH = "./src/static"

func main() {

	// Init DB if it doesn't exist
	if !utils.FileExists("/tmp/minitwit.db") {
		fmt.Println("Initializing database")
		database.InitDB(DATABASE)
	}
	fmt.Println("Running: localhost:5000/public")
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)
	gob.Register(&types.User{})

	r.HandleFunc("/", authentication.Auth(timeline))
	r.HandleFunc("/public", publicTimeline)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/addMessage", authentication.Auth(addMessage)).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	http.ListenAndServe(":5000", r)

}

func getUserID(username string) (int, error) {
	//Convenience method to loop up the id for a username
	var id int
	var rows *sql.Rows = database.QueryDB("select user_id from user where username = ?", utils.StringArrayToInterfaceArray([]string{username}), DATABASE)
	for rows.Next() {
		err := rows.Scan(&id)

		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

func format_datetime(timestamp string) string {
	splittedTimestamp := strings.Split(timestamp, ".")
	time := splittedTimestamp[0]
	return time[:16]
}

func gravatarURL(email string, size int) string {
	//Return the gravatar image for the given email address
	return email
}

func timeline(w http.ResponseWriter, r *http.Request) {
	user_ := authentication.GetSessionValue(w, r, "user")
	user := user_.(*(types.User))
	user_id := user.User_id
	stmt, err := DATABASE.Prepare(`select message.*, user.* from message, user
	where message.flagged = 0 and message.author_id = user.user_id and (
		user.user_id = ? or
		user.user_id in (select whom_id from follower
								where who_id = ?))
	order by message.pub_date desc limit ?`)
	rows, err := stmt.Query(user_id, user_id, PER_PAGE)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusForbidden)
	}

	messages := []types.MessageViewData{}

	for rows.Next() {
		var message_id int
		var author_id int
		var text string
		var pub_date string
		var flagged int

		var user_id int
		var username string
		var email string
		var pw_hash string

		err = rows.Scan(&message_id, &author_id, &text, &pub_date, &flagged, &user_id, &username, &email, &pw_hash)
		if err != nil {
			log.Fatal(err)
		}

		message := types.MessageViewData{
			Text:        text,
			Email:       email,
			GravatarURL: gravatarURL(email, 64),
			Username:    username,
			Pub_date:    format_datetime(pub_date),
		}
		messages = append(messages, message)
	}

	data := types.RequestData{
		RequestEndpoint: "",
		Title:           "title",
		Messages:        messages,
		IsLoggedIn:      true,
		SessionUser:     user.Username,
		UserProfile:     "",
	}
	utils.RenderTemplate(w, utils.TIMELINE, data)
}

func publicTimeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: " + r.RemoteAddr)

	stmt, err := DATABASE.Prepare(`select message.*, user.* from message, user
	where message.flagged = 0 and message.author_id = user.user_id
	order by message.pub_date desc limit ?`)
	rows, err := stmt.Query(PER_PAGE)

	messages := []types.MessageViewData{}

	for rows.Next() {
		var message_id int
		var author_id int
		var text string
		var pub_date string
		var flagged int

		var user_id int
		var username string
		var email string
		var pw_hash string

		err = rows.Scan(&message_id, &author_id, &text, &pub_date, &flagged, &user_id, &username, &email, &pw_hash)
		if err != nil {
			log.Fatal(err)
		}

		message := types.MessageViewData{
			Text:        text,
			Email:       email,
			GravatarURL: gravatarURL(email, 64),
			Username:    username,
			Pub_date:    format_datetime(pub_date),
		}

		messages = append(messages, message)
	}

	data := types.RequestData{
		Title:           "MEGA TITLE",
		RequestEndpoint: "timeline",
		Messages:        messages,
	}

	user := authentication.GetSessionValue(w, r, "user")
	if user != nil {
		username := (user.(*types.User)).Username
		data.IsLoggedIn = true
		data.SessionUser = username
	}

	utils.RenderTemplate(w, utils.TIMELINE, data)
}

func userTimeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("User timeline hit")

	user := authentication.GetSessionValue(w, r, "user")
	username := (user.(*types.User)).Username
	user_id := (user.(*types.User)).User_id
	// cookie, _ := r.Cookie(username)
	// fmt.Println(w, username)
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

	stmt, err = DATABASE.Prepare(`select message.*, user.* from message, user
	where user.user_id = message.author_id and user.user_id = ? 
	order by message.pub_date desc limit ?`)

	rows, err := stmt.Query(profile_user_id, PER_PAGE)

	messages := []types.MessageViewData{}
	for rows.Next() {
		var message_id int
		var author_id int
		var text string
		var pub_date string
		var flagged int

		var user_id int
		var username string
		var email string
		var pw_hash string

		err = rows.Scan(&message_id, &author_id, &text, &pub_date, &flagged, &user_id, &username, &email, &pw_hash)

		if err != nil {
			fmt.Println(err)
		}

		message := types.MessageViewData{
			Text:        text,
			Email:       email,
			GravatarURL: gravatarURL(email, 64),
			Username:    username,
			Pub_date:    format_datetime(pub_date),
		}
		messages = append(messages, message)
	}

	data := types.RequestData{
		Title:           "title",
		RequestEndpoint: "userTimeline",
		Messages:        messages,
		IsLoggedIn:      true,
		SessionUser:     username,
		UserProfile:     profile_username,
		Followed:        followed,
	}

	utils.RenderTemplate(w, utils.TIMELINE, data)
}

func followUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "follow_user hit")
	username := r.FormValue("username")
	userID, err := getUserID(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).User_id
	queryString := `INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`
	statement, err := DATABASE.Prepare(queryString)
	_, err = statement.Exec(sessionUserID, userID)
	checkErr(err)

	authentication.Flash(w, r, "You are now following "+username)

	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func unfollowUser(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "unfollow_user")
	username := r.FormValue("username")
	userID, err := getUserID(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).User_id
	queryString := `DELETE FROM follower WHERE who_id = ? AND whom_id = ?`
	statement, err := DATABASE.Prepare(queryString)
	_, err = statement.Exec(sessionUserID, userID)
	checkErr(err)
	authentication.Flash(w, r, "You are no longer following "+username)

	http.Redirect(w, r, "/", http.StatusFound)
}

func addMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println(w, "addMessage hit")

	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	user_id := (user.(*types.User)).User_id

	if text != "" {
		queryString := `INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`
		statement, err := DATABASE.Prepare(queryString)
		_, err = statement.Exec(user_id, text, time.Now())
		checkErr(err)
	}

	http.Redirect(w, r, "/public", http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		loginGet(w, r)
	} else if r.Method == "POST" {
		loginPost(w, r)
	}

}

func loginGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.LOGIN, nil)
}

func loginPost(w http.ResponseWriter, r *http.Request) {
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

	userFound, user := authenticate(username, password)

	if !userFound {
		data.HasError = true
		data.ErrorMsg = "Cannot recognize user"
		data.IsLoggedIn = false
		utils.RenderTemplate(w, utils.LOGIN, data)
		return
	}
	err := authentication.PutSessionValue(w, r, "user", user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		data.IsLoggedIn = true
		data.Username = user.Username
		data.RequestEndpoint = ""
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func authenticate(username string, password string) (bool, *types.User) {

	stmt, err := DATABASE.Prepare("SELECT * FROM user WHERE username = ?")
	user_id := -999
	email := ""
	pw_hash := ""
	err = stmt.QueryRow(username).Scan(&user_id, &username, &email, &pw_hash)

	user := &types.User{
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
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
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
	utils.RenderTemplate(w, utils.REGISTER, nil)
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
		data := struct {
			HasError   bool
			ErrorMsg   string
			IsLoggedIn bool
		}{true, errorMsg, false}
		utils.RenderTemplate(w, utils.REGISTER, data)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}

}
