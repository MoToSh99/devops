package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	authentication "github.com/matt035343/devops/src/authentication"
	"github.com/matt035343/devops/src/types"
	"github.com/matt035343/devops/src/utils"
	"golang.org/x/crypto/bcrypt"
)

var PER_PAGE = 30
var STATIC_ROOT_PATH = "./src/static"
var LATEST int64 = 0

func (s *Server) getUserID(username string) (int, error) {
	u, err := s.db.GetUserFromUsername(username)
	if err != nil {
		return 0, err
	}
	return u.UserID, nil
}

func (s *Server) timeline(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user").(*(types.User))

	messages, err := s.db.GetTimelineViewMessages(user.UserID, PER_PAGE)
	if err != nil {
		panic(err)
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

func (s *Server) publicTimeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: " + r.RemoteAddr)

	messages, err := s.db.GetPublicViewMessages(PER_PAGE)
	if err != nil {
		panic(err)
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

func (s *Server) userTimeline(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	profile, err := s.db.GetUserFromUsername(username)
	if err != nil {
		panic(err)
	}

	user := authentication.GetSessionValue(w, r, "user")
	userID := (user.(*types.User)).UserID
	follower, err := s.db.GetFollower(userID, profile.UserID)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}

	messages, err := s.db.GetUserViewMessages(profile.UserID, PER_PAGE)
	if err != nil {
		panic(err)
	}

	data := types.RequestData{
		Title:           "title",
		RequestEndpoint: "userTimeline",
		Messages:        messages,
		IsLoggedIn:      true,
		SessionUser:     (user.(*types.User)).Username,
		UserProfile:     profile.Username,
		Followed:        follower.IsValidRelation(),
	}

	utils.RenderTemplate(w, utils.TIMELINE, data)
}

func (s *Server) followUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := s.db.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = s.db.AddFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	authentication.Flash(w, r, "You are now following "+username)

	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func (s *Server) unfollowUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := s.db.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = s.db.DeleteFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	authentication.Flash(w, r, "You are no longer following "+username)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *Server) AddMessage(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	userID := (user.(*types.User)).UserID

	if text != "" {
		err := s.db.AddMessage(userID, text, time.Now())
		if err != nil {
			panic(err)
		}
	}

	http.Redirect(w, r, "/public", http.StatusFound)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.loginGet(w, r)
	} else if r.Method == "POST" {
		s.loginPost(w, r)
	}

}

func (s *Server) loginGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.LOGIN, nil)
}

func (s *Server) loginPost(w http.ResponseWriter, r *http.Request) {
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

	userFound, user := s.authenticate(username, password)

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

func (s *Server) authenticate(username string, password string) (bool, *types.User) {
	user, err := s.db.GetUserFromUsername(username)
	if err != nil {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, &user
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/public", http.StatusFound)

}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.registerGet(w, r)
	} else if r.Method == "POST" {
		s.registerPost(w, r)
	}

}

func (s *Server) isUsernameAvailable(username string) bool {
	_, err := s.db.GetUserFromUsername(username)
	if err != nil {
		return true
	}
	return false
}

func (s *Server) registerGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.REGISTER, nil)
}

func (s *Server) registerUser(username string, email string, hashedPassword string) bool {
	err := s.db.AddUser(username, email, hashedPassword)
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

func (s *Server) registerPost(w http.ResponseWriter, r *http.Request) {
	username_from_form := r.FormValue("username")
	errorMsg := ""
	decoder := json.NewDecoder(r.Body)
	var registerRequest types.RegisterRequest
	decoder.Decode(&registerRequest)
	if registerRequest != (types.RegisterRequest{}) && username_from_form == "" {
		s.registerPostFromJson(w, r, registerRequest)
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
	} else if !s.isUsernameAvailable(r.FormValue("username")) {
		errorMsg = utils.USERNAME_TAKEN
	} else {
		hashedPasswordInBytes, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 14)
		s.registerUser(r.FormValue("username"), r.FormValue("email"), string(hashedPasswordInBytes))
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

func (s *Server) registerPostFromJson(w http.ResponseWriter, r *http.Request, registerRequest types.RegisterRequest) {
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
	} else if !s.isUsernameAvailable(registerRequest.Username) {
		error = utils.USERNAME_TAKEN
	}
	if error != "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(types.ErrorMsgResponse{Status: 400, Error_msg: error})
		return
	} else {
		res := s.registerUser(registerRequest.Username, registerRequest.Email, registerRequest.Pwd)
		if res {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

func (s *Server) getUserIDFromUrl(r *http.Request) (int, error) {
	username := mux.Vars(r)["username"]
	u, err := s.db.GetUserFromUsername(username)
	return u.UserID, err
}

// func tweetsGet(w http.ResponseWriter, r *http.Request) {
// 	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
// 	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
// 	if latest_err != nil || no_msgs_err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	} else {
// 		LATEST = latest
// 		messages := database.QueryMessages(`select message.*, user.* from message, user
// 			where message.flagged = 0 and message.author_id = user.user_id
// 			order by message.pub_date desc limit ?`, no_msgs)

// 		filtered_msgs := types.ConvertToTweetResponse(messages)

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(filtered_msgs)
// 	}
// }

// func tweetsUsername(w http.ResponseWriter, r *http.Request) {
// 	if r.Method == "GET" {
// 		tweetsUsernameGet(w, r)
// 	} else if r.Method == "POST" {
// 		tweetsUsernamePost(w, r)
// 	}
// }

// func tweetsUsernameGet(w http.ResponseWriter, r *http.Request) {
// 	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
// 	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

// 	if latest_err != nil || no_msgs_err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	userID, userID_err := getUserIDFromUrl(r)
// 	if userID_err != nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	} else {

// 		LATEST = latest
// 		messages := database.QueryMessages(`select message.*, user.* from message, user
// 			where message.flagged = 0 and message.author_id = user.user_id and user.user_id = ?
// 			order by message.pub_date desc limit ?`, userID, no_msgs)

// 		filtered_msgs := types.ConvertToTweetResponse(messages)

// 		w.Header().Set("Content-Type", "application/json")
// 		json.NewEncoder(w).Encode(filtered_msgs)
// 	}
// }

// func tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
// 	userID, userID_err := getUserIDFromUrl(r)
// 	if userID_err != nil {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}

// 	decoder := json.NewDecoder(r.Body)
// 	var tweet types.TweetRequest
// 	tweet_err := decoder.Decode(&tweet)
// 	if tweet_err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	} else {
// 		err := database.AlterDB(`INSERT INTO message (author_id, text, pub_date, flagged) VALUES (?, ?, ?, 0)`, userID, tweet.Content, time.Now())
// 		if err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

// func followUsername(w http.ResponseWriter, r *http.Request) {
// 	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

// 	if latest_err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	LATEST = latest
// 	userID, userID_err := getUserIDFromUrl(r)
// 	if userID_err != nil || userID == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}

// 	if r.Method == "GET" {
// 		followUsernameGet(w, r, userID)
// 	} else if r.Method == "POST" {
// 		decoder := json.NewDecoder(r.Body)
// 		var followRequest types.FollowRequest
// 		decoder.Decode(&followRequest)
// 		if followRequest.Follow == "" && followRequest.Unfollow == "" {
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}

// 		isFollow := followRequest.Follow != ""
// 		if isFollow {
// 			followUsernamePost(w, r, userID, followRequest)
// 		} else {
// 			unFollowUsernamePost(w, r, userID, followRequest)
// 		}
// 	}
// }

// func followUsernameGet(w http.ResponseWriter, r *http.Request, userID int) {
// 	no_followers, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
// 	if no_msgs_err != nil {
// 		w.WriteHeader(http.StatusBadRequest)
// 		return
// 	}
// 	userID, userID_err := getUserIDFromUrl(r)
// 	if userID_err != nil || userID == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	}
// 	followers := database.QueryFollowers(`SELECT user.username FROM user
// 										INNER JOIN follower ON follower.whom_id=user.user_id
// 										WHERE follower.who_id=?
// 										LIMIT ?`, userID, no_followers)
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(followers)
// }

// func followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
// 	follows_userID, follows_user_err := getUserID(followRequest.Follow)
// 	if follows_user_err != nil || follows_userID == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	} else {
// 		follow_insert_err := database.AlterDB(`INSERT INTO follower (who_id, whom_id) VALUES (?, ?)`, userID, follows_userID)
// 		if follow_insert_err != nil {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

// func unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
// 	unfollows_userID, unfollows_user_err := getUserID(unfollowRequest.Unfollow)
// 	if unfollows_user_err != nil || unfollows_userID == 0 {
// 		w.WriteHeader(http.StatusNotFound)
// 		return
// 	} else {
// 		unfollow_err := database.AlterDB(`DELETE FROM follower WHERE who_id = ? AND whom_id = ?`, userID, unfollows_userID)
// 		if unfollow_err != unfollow_err {
// 			w.WriteHeader(http.StatusInternalServerError)
// 			return
// 		}
// 		w.WriteHeader(http.StatusNoContent)
// 	}
// }

// func latest(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(types.LatestResponse{Latest: LATEST})
// }
