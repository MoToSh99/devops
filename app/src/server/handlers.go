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
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"
	"golang.org/x/crypto/bcrypt"
)

var perPage = 30
var staticRootPath = "./src/static"
var latest int64 = 0

func (s *Server) getUserID(username string) (int, error) {
	u, err := s.db.GetUserFromUsername(username)
	if err != nil {
		return 0, err
	}
	return u.UserID, nil
}

func (s *Server) timeline(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user").(*(types.User))

	messages, err := s.db.GetTimelineViewMessages(user.UserID, perPage)
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
	utils.RenderTemplate(w, utils.Timeline, data)
}

func (s *Server) publicTimeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: " + r.RemoteAddr)

	messages, err := s.db.GetPublicViewMessages(perPage)
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

	utils.RenderTemplate(w, utils.Timeline, data)
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

	messages, err := s.db.GetUserViewMessages(profile.UserID, perPage)
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

	utils.RenderTemplate(w, utils.Timeline, data)
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
	middleware.UsersFollowed.Inc()
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
	middleware.UsersUnfollowed.Inc()
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
		middleware.MessagesSent.Inc()
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
	utils.RenderTemplate(w, utils.Login, nil)
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
		utils.RenderTemplate(w, utils.Login, data)
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
	utils.RenderTemplate(w, utils.Register, nil)
}

func (s *Server) registerUser(username string, email string, hashedPassword string) bool {
	err := s.db.AddUser(username, email, hashedPassword)
	if err != nil {
		return false
	}
	middleware.UsersRegistered.Inc()
	return true
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Server) registerPost(w http.ResponseWriter, r *http.Request) {
	usernameFromForm := r.FormValue("username")
	errorMsg := ""
	decoder := json.NewDecoder(r.Body)
	var registerRequest types.RegisterRequest
	decoder.Decode(&registerRequest)
	if registerRequest != (types.RegisterRequest{}) && usernameFromForm == "" {
		s.registerPostFromJson(w, r, registerRequest)
		return
	}

	if usernameFromForm == "" {
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
		utils.RenderTemplate(w, utils.Register, data)

	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func (s *Server) registerPostFromJson(w http.ResponseWriter, r *http.Request, registerRequest types.RegisterRequest) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 64)
	if latestErr != nil {
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
		json.NewEncoder(w).Encode(types.ErrorMsgResponse{Status: 400, ErrorMsg: error})
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
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		s.db.SetLatest(latest)
		messages, err := s.db.GetPublicViewMessages(int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
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
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)
	noMsgs, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)

	if latestErr != nil || noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {

		s.db.SetLatest(latest)
		messages, err := s.db.GetUserViewMessages(userID, int(noMsgs))
		if err != nil {
			panic(err)
		}

		filteredMsgs := types.ConvertToTweetResponse(messages)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(filteredMsgs)
	}
}

func (s *Server) tweetsUsernamePost(w http.ResponseWriter, r *http.Request) {
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var tweet types.TweetRequest
	tweetErr := decoder.Decode(&tweet)
	if tweetErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	} else {
		err := s.db.AddMessage(userID, tweet.Content, time.Now())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.MessagesSent.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) followUsername(w http.ResponseWriter, r *http.Request) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 32)

	if latestErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.db.SetLatest(latest)
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil || userID == 0 {
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
	noFollowers, noMsgsErr := strconv.ParseInt(r.URL.Query().Get("no"), 10, 64)
	if noMsgsErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	userID, userIDErr := s.getUserIDFromUrl(r)
	if userIDErr != nil || userID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	followers, err := s.db.GetFollowers(userID, int(noFollowers))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(followers)
}

func (s *Server) followUsernamePost(w http.ResponseWriter, r *http.Request, userID int, followRequest types.FollowRequest) {
	followsUserID, followsUserErr := s.getUserID(followRequest.Follow)
	if followsUserErr != nil || followsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		followInsertErr := s.db.AddFollower(userID, followsUserID)
		if followInsertErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersFollowed.Inc()
		w.WriteHeader(http.StatusNoContent)
	}
}

func (s *Server) unFollowUsernamePost(w http.ResponseWriter, r *http.Request, userID int, unfollowRequest types.FollowRequest) {
	unfollowsUserID, unfollowsUserErr := s.getUserID(unfollowRequest.Unfollow)
	if unfollowsUserErr != nil || unfollowsUserID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return
	} else {
		unfollowErr := s.db.DeleteFollower(userID, unfollowsUserID)
		if unfollowErr != unfollowErr {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		middleware.UsersUnfollowed.Inc()
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
