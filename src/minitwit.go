package main

import (
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	authentication "go/src/authentication"
	"go/src/database"
	"go/src/types"
	"go/src/utils"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var PER_PAGE = 30
var STATIC_ROOT_PATH = "./src/static"
var LATEST int64 = 0

func main() {

	// Init DB if it doesn't exist
	if !utils.FileExists("/tmp/minitwit.db") {
		fmt.Println("Initializing database")
		database.InitDB()
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
	r.HandleFunc("/msgs", tweetGet).Methods("Get")
	r.HandleFunc("/msgs/{username}", tweet).Methods("POST")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	http.ListenAndServe(":5000", r)

}

func getUserID(username string) (int, error) {
	//Convenience method to loop up the id for a username
	var id int
	var rows *sql.Rows = database.QueryRowsDB("select user_id from user where username = ?", username)
	for rows.Next() {
		err := rows.Scan(&id)

		if err != nil {
			return 0, err
		}
	}
	return id, nil
}

func timeline(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user").(*(types.User))
	userID := user.UserID

	messages := database.QueryMessages(`select message.*, user.* from message, user
	where message.flagged = 0 and message.author_id = user.user_id and (
		user.user_id = ? or
		user.user_id in (select whom_id from follower
								where who_id = ?))
	order by message.pub_date desc limit ?`, userID, userID, PER_PAGE)

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

	messages := database.QueryMessages(`select message.*, user.* from message, user
	where message.flagged = 0 and message.author_id = user.user_id
	order by message.pub_date desc limit ?`, PER_PAGE)

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

	username := mux.Vars(r)["username"]
	profile := types.User{}

	err := database.QueryRowDB("SELECT * FROM user WHERE username = ?", username).Scan(&profile.UserID, &profile.Username, &profile.Email, &profile.PasswordHash)
	if err != nil && err == sql.ErrNoRows {
		panic(err)
	}
	user := authentication.GetSessionValue(w, r, "user")
	userID := (user.(*types.User)).UserID
	follower := types.Follower{}
	err = database.QueryRowDB(`select 1 from follower where
	follower.who_id = ? and follower.whom_id = ?`, userID, profile.UserID).Scan(&follower.WhoID, &follower.WhomID)
	if err != nil && err == sql.ErrNoRows {
		follower.Followed = false
	} else {
		follower.Followed = true
	}

	messages := database.QueryMessages(`select message.*, user.* from message, user
	where user.user_id = message.author_id and user.user_id = ? 
	order by message.pub_date desc limit ?`, profile.UserID, PER_PAGE)

	data := types.RequestData{
		Title:           "title",
		RequestEndpoint: "userTimeline",
		Messages:        messages,
		IsLoggedIn:      true,
		SessionUser:     (user.(*types.User)).Username,
		UserProfile:     profile.Username,
		Followed:        follower.Followed,
	}

	utils.RenderTemplate(w, utils.TIMELINE, data)
}

func followUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	userID, err := getUserID(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = database.AlterDB(`INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`, sessionUserID, userID)
	checkErr(err)
	authentication.Flash(w, r, "You are now following "+username)

	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func unfollowUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	userID, err := getUserID(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = database.AlterDB(`DELETE FROM follower WHERE who_id = ? AND whom_id = ?`, sessionUserID, userID)
	checkErr(err)
	authentication.Flash(w, r, "You are no longer following "+username)

	http.Redirect(w, r, "/", http.StatusFound)
}

func addMessage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("addMessage hit")

	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	userID := (user.(*types.User)).UserID

	if text != "" {
		err := database.AlterDB(`INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`, userID, text, time.Now())
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
	user := &types.User{}

	err := database.QueryRowDB("SELECT * FROM user WHERE username = ?", username).Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash)

	if err != nil && err == sql.ErrNoRows {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
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

func isUsernameAvailable(username string) bool {
	user := types.User{}
	err := database.QueryRowDB("SELECT * FROM user WHERE username = ?", username).Scan(&user.UserID, &user.Username, &user.Email, &user.PasswordHash)
	if err != nil {
		return true
	}
	return false
}

func registerGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.REGISTER, nil)
}

func registerUser(username string, email string, hashedPassword string) bool {
	err := database.AlterDB("INSERT INTO user (username, email, pw_hash) VALUES (?, ?, ?)", username, email, hashedPassword)
	if err != nil {
		return false
	}
	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func registerPost(w http.ResponseWriter, r *http.Request) *http.Response {
	errorMsg := ""
	fmt.Println(r.Form)
	decoder := json.NewDecoder(r.Body)
	var registerRequest types.RegisterRequest
	decoder.Decode(&registerRequest)
	if registerRequest != (types.RegisterRequest{}) {
		registerPostFromJson(w, r, registerRequest)
	}
	fmt.Println("qawd")
	fmt.Println(r.FormValue("username"))
	if r.FormValue("username") == "" {
		errorMsg = "You have to enter a username"
	} else if r.FormValue("email") == "" || !strings.Contains(r.FormValue("email"), "@") {
		errorMsg = "You have to enter a valid email address"
	} else if r.FormValue("password") == "" {
		errorMsg = "You have to enter a password"
	} else if r.FormValue("password") != r.FormValue("password2") {
		errorMsg = "The two passwords do not match"
	} else if !isUsernameAvailable(r.FormValue("username")) {
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
		return r.Response
	} else {
		fmt.Println("else")
		http.Redirect(w, r, "/login", http.StatusNoContent)
		return r.Response
	}
}

func registerPostFromJson(w http.ResponseWriter, r *http.Request, registerRequest types.RegisterRequest) {
	var res = registerUser(registerRequest.Username, registerRequest.Email, registerRequest.Pwd)
	if res {
		http.Redirect(w, r, "/login", http.StatusNoContent)
	}
}

func tweetGet(w http.ResponseWriter, r *http.Request) {
	latest, err1 := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	no_msgs, err2 := strconv.ParseInt(r.URL.Query().Get("no_msgs"), 10, 64)
	if err1 != nil || err2 != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		LATEST = latest
		messages := database.QueryMessages(`select message.*, user.* from message, user
			where message.flagged = 0 and message.author_id = user.user_id
			order by message.pub_date desc limit ?`, no_msgs)

		var filtered_msgs []types.TweetResponse

		for _, msg := range messages {
			filtered_msgs = append(filtered_msgs, types.TweetResponse{Content: msg.Text, User: msg.Username, Pub_date: msg.PublishedDate})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filtered_msgs)

	}
}

func tweet(w http.ResponseWriter, r *http.Request) {
	// username := mux.Vars(r)["username"]
	// decoder := json.NewDecoder(r.Body)
	// var tweet types.TweetRequest
	// err := decoder.Decode(&tweet)
	// if err != nil {
	// 	panic(err)
	// }
	// http.Redirect(w, r, "aa", http.StatusNoContent)
	// // fmt.Println(tweet.Content)
}
