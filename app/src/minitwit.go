package main

import (
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	authentication "github.com/matt035343/devops/src/authentication"
	"github.com/matt035343/devops/src/database"
	"github.com/matt035343/devops/src/types"
	"github.com/matt035343/devops/src/utils"

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
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)
	gob.Register(&types.User{})

	r.HandleFunc("/", authentication.Auth(timeline))
	r.HandleFunc("/public", publicTimeline)
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/addMessage", authentication.Auth(AddMessage)).Methods("POST")
	r.HandleFunc("/login", Login).Methods("GET", "POST")
	r.HandleFunc("/register", Register).Methods("GET", "POST")
	r.HandleFunc("/msgs", tweetsGet).Methods("Get")
	r.HandleFunc("/msgs/{username}", tweetsUsername).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", followUsername).Methods("GET", "POST")
	r.HandleFunc("/latest", latest).Methods("GET")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))

	fmt.Println("Running: localhost:5000/public")
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

func AddMessage(w http.ResponseWriter, r *http.Request) {

	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	userID := (user.(*types.User)).UserID

	if text != "" {
		err := database.AlterDB(`INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`, userID, text, time.Now())
		checkErr(err)
	}

	http.Redirect(w, r, "/public", http.StatusFound)
}

func Login(w http.ResponseWriter, r *http.Request) {
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
		data.ErrorMsg = "Invalid password"
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

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, user
}

func Logout(w http.ResponseWriter, r *http.Request) {
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/public", http.StatusFound)

}

func Register(w http.ResponseWriter, r *http.Request) {
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

func registerPost(w http.ResponseWriter, r *http.Request) {
	username_from_form := r.FormValue("username")
	errorMsg := ""
	decoder := json.NewDecoder(r.Body)
	var registerRequest types.RegisterRequest
	decoder.Decode(&registerRequest)
	if registerRequest != (types.RegisterRequest{}) && username_from_form == "" {
		registerPostFromJson(w, r, registerRequest)
		return
	}

	if username_from_form == "" {
		errorMsg = utils.ENTER_A_USERNAME
	} else if r.FormValue("email") == "" || !strings.Contains(r.FormValue("email"), "@") {
		errorMsg = utils.ENTER_A_VALID_EMAIL
	} else if r.FormValue("password") == "" {
		errorMsg = utils.YOU_HAVE_TO_ENTER_A_PASSWORD
	} else if r.FormValue("password") != r.FormValue("password2") {
		errorMsg = utils.PASSWORDS_DOES_NOT_MATCH_MESSAGE
	} else if !isUsernameAvailable(r.FormValue("username")) {
		errorMsg = utils.USERNAME_TAKEN
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

func registerPostFromJson(w http.ResponseWriter, r *http.Request, registerRequest types.RegisterRequest) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 64)
	if latest_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	LATEST = latest
	error := ""
	if registerRequest.Username == "" {
		error = utils.ENTER_A_USERNAME
	} else if registerRequest.Email == "" || !strings.Contains(registerRequest.Email, "@") {
		error = utils.ENTER_A_VALID_EMAIL
	} else if registerRequest.Pwd == "" {
		error = utils.YOU_HAVE_TO_ENTER_A_PASSWORD
	} else if !isUsernameAvailable(registerRequest.Username) {
		error = utils.USERNAME_TAKEN
	}
	if error != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ErrorMsgResponse{Status: 400, Error_msg: error})
		return
	} else {
		res := registerUser(registerRequest.Username, registerRequest.Email, registerRequest.Pwd)
		if res {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func getUserIDFromUrl(r *http.Request) (int, error) {
	username := mux.Vars(r)["username"]
	return getUserID(username)
}

func tweetsGet(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if latest_err != nil || no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		LATEST = latest
		messages := database.QueryMessages(`select message.*, user.* from message, user
			where message.flagged = 0 and message.author_id = user.user_id
			order by message.pub_date desc limit ?`, no_msgs)

		filtered_msgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filtered_msgs)
	}
}

func tweetsUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tweetsUsernameGet(w, r)
	} else if r.Method == "POST" {
		tweetsUsernamePost(w, r)
	}
}

func tweetsUsernameGet(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

	if latest_err != nil || no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userID_err := getUserIDFromUrl(r)
	if userID_err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		LATEST = latest
		messages := database.QueryMessages(`select message.*, user.* from message, user
			where message.flagged = 0 and message.author_id = user.user_id and user.user_id = ?
			order by message.pub_date desc limit ?`, userID, no_msgs)

		filtered_msgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filtered_msgs)
	}
}

func tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
	userID, userID_err := getUserIDFromUrl(r)
	if userID_err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var tweet types.TweetRequest
	tweet_err := decoder.Decode(&tweet)
	if tweet_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		err := database.AlterDB(`INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`, userID, tweet.Content, time.Now())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func followUsername(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

	if latest_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	LATEST = latest
	userID, userID_err := getUserIDFromUrl(r)
	if userID_err != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		followUsernameGet(w, r, userID)
	} else if r.Method == "POST" {
		decoder := json.NewDecoder(r.Body)
		var followRequest types.FollowRequest
		decoder.Decode(&followRequest)
		if followRequest.Follow == "" && followRequest.Unfollow == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		isFollow := followRequest.Follow != ""
		if isFollow {
			followUsernamePost(w, r, userID, followRequest)
		} else {
			unFollowUsernamePost(w, r, userID, followRequest)
		}
	}
}

func followUsernameGet(w http.ResponseWriter, r *http.Request, userID int) {
	no_followers, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userID_err := getUserIDFromUrl(r)
	if userID_err != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	followers := database.QueryFollowers(`SELECT user.username FROM user
										INNER JOIN follower ON follower.whom_id=user.user_id
										WHERE follower.who_id=?
										LIMIT ?`, userID, no_followers)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
	follows_userID, follows_user_err := getUserID(followRequest.Follow)
	if follows_user_err != nil || follows_userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		follow_insert_err := database.AlterDB(`INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`, userID, follows_userID)
		if follow_insert_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
	unfollows_userID, unfollows_user_err := getUserID(unfollowRequest.Unfollow)
	if unfollows_user_err != nil || unfollows_userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		unfollow_err := database.AlterDB(`DELETE FROM follower WHERE who_id = ? AND whom_id = ?`, userID, unfollows_userID)
		if unfollow_err != unfollow_err {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func latest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.LatestResponse{Latest: LATEST})
}
