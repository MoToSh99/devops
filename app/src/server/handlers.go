package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	authentication "github.com/matt035343/devops/app/src/authentication"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"
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
	if err != nil && err != gorm.ErrRecordNotFound {
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

func (s *Server) addMessage(w http.ResponseWriter, r *http.Request) {
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

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
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
	if err != nil && err != gorm.ErrRecordNotFound {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, &user
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/public", http.StatusFound)

}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
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
	s.db.SetLatest(latest)
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

func (s *Server) tweetsGet(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if latest_err != nil || no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		s.db.SetLatest(latest)
		messages, err := s.db.GetPublicViewMessages(int(no_msgs))
		if err != nil {
			panic(err)
		}

		filtered_msgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filtered_msgs)
	}
}

func (s *Server) tweetsUsername(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.tweetsUsernameGet(w, r)
	} else if r.Method == "POST" {
		s.tweetsUsernamePost(w, r)
	}
}

func (s *Server) tweetsUsernameGet(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	no_msgs, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

	if latest_err != nil || no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userID_err := s.getUserIDFromUrl(r)
	if userID_err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		s.db.SetLatest(latest)
		messages, err := s.db.GetUserViewMessages(userID, int(no_msgs))
		if err != nil {
			panic(err)
		}

		filtered_msgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filtered_msgs)
	}
}

func (s *Server) tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
	userID, userID_err := s.getUserIDFromUrl(r)
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
		err := s.db.AddMessage(userID, tweet.Content, time.Now())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) followUsername(w http.ResponseWriter, r *http.Request) {
	latest, latest_err := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

	if latest_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.db.SetLatest(latest)
	userID, userID_err := s.getUserIDFromUrl(r)
	if userID_err != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if r.Method == "GET" {
		s.followUsernameGet(w, r, userID)
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
			s.followUsernamePost(w, r, userID, followRequest)
		} else {
			s.unFollowUsernamePost(w, r, userID, followRequest)
		}
	}
}

func (s *Server) followUsernameGet(w http.ResponseWriter, r *http.Request, userID int) {
	no_followers, no_msgs_err := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if no_msgs_err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userID_err := s.getUserIDFromUrl(r)
	if userID_err != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	followers, err := s.db.GetFollowers(userID, int(no_followers))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (s *Server) followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
	follows_userID, follows_user_err := s.getUserID(followRequest.Follow)
	if follows_user_err != nil || follows_userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		follow_insert_err := s.db.AddFollower(userID, follows_userID)
		if follow_insert_err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
	unfollows_userID, unfollows_user_err := s.getUserID(unfollowRequest.Unfollow)
	if unfollows_user_err != nil || unfollows_userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		unfollow_err := s.db.DeleteFollower(userID, unfollows_userID)
		if unfollow_err != unfollow_err {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) latest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	l, err := s.db.GetLatest()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(l)
}
