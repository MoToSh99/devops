package client

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

func (s *ClientServer) timeline(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user").(*(types.User))

	messages, err := s.DB.GetTimelineViewMessages(user.UserID, perPage)
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
		HasError:        false,
		ErrorMsg:        "",
	}
	utils.RenderTemplate(w, utils.Timeline, data)
}

func (s *ClientServer) publicTimeline(w http.ResponseWriter, r *http.Request) {
	fmt.Println("We got a visitor from: " + r.RemoteAddr)

	messages, err := s.DB.GetPublicViewMessages(perPage)
	if err != nil {
		panic(err)
	}

	data := types.RequestData{
		Title:           "MEGA TITLE",
		RequestEndpoint: "timeline",
		Messages:        messages,
		HasError:        false,
		ErrorMsg:        "",
	}

	user := authentication.GetSessionValue(w, r, "user")
	if user != nil {
		username := (user.(*types.User)).Username
		data.IsLoggedIn = true
		data.SessionUser = username
	}

	utils.RenderTemplate(w, utils.Timeline, data)
}

func (s *ClientServer) userTimeline(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	profile, err := s.DB.GetUserFromUsername(username)
	user := authentication.GetSessionValue(w, r, "user")

	userID := (user.(*types.User)).UserID

	if err != nil && err == gorm.ErrRecordNotFound {
		error_response := types.RequestData{
			Title:           "title",
			RequestEndpoint: "userTimeline",
			Messages:        nil,
			IsLoggedIn:      true,
			SessionUser:     (user.(*types.User)).Username,
			UserProfile:     profile.Username,
			Followed:        false,
			HasError:        true,
			ErrorMsg:        "Sorry! We could not find the user.",
		}
		utils.RenderTemplate(w, utils.Timeline, error_response)
		return
	} else if err != nil {
		panic(err)
	}

	follower, err := s.DB.GetFollower(userID, profile.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		panic(err)
	}

	messages, err := s.DB.GetUserViewMessages(profile.UserID, perPage)
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
		HasError:        false,
		ErrorMsg:        "",
	}

	utils.RenderTemplate(w, utils.Timeline, data)
}

func (s *ClientServer) followUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := s.DB.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = s.DB.AddFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	middleware.UsersFollowed.Inc()
	authentication.Flash(w, r, "You are now following "+username)

	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func (s *ClientServer) unfollowUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := s.DB.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = s.DB.DeleteFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	middleware.UsersUnfollowed.Inc()
	authentication.Flash(w, r, "You are no longer following "+username)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (s *ClientServer) addMessage(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	userID := (user.(*types.User)).UserID

	if text != "" {
		err := s.DB.AddMessage(userID, text, time.Now())
		if err != nil {
			panic(err)
		}
		middleware.MessagesSent.Inc()
	}

	http.Redirect(w, r, "/public", http.StatusFound)
}

func (s *ClientServer) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.loginGet(w, r)
	} else if r.Method == "POST" {
		s.loginPost(w, r)
	}

}

func (s *ClientServer) loginGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.Login, nil)
}

func (s *ClientServer) loginPost(w http.ResponseWriter, r *http.Request) {
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

func (s *ClientServer) authenticate(username string, password string) (bool, *types.User) {
	user, err := s.DB.GetUserFromUsername(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, &user
}

func (s *ClientServer) logout(w http.ResponseWriter, r *http.Request) {
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/public", http.StatusFound)

}

//Register Handler for the register endpoint
func (s *ClientServer) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		s.registerGet(w, r)
	} else if r.Method == "POST" {
		s.registerPost(w, r)
	}

}

func (s *ClientServer) isUsernameAvailable(username string) bool {
	_, err := s.DB.GetUserFromUsername(username)
	return err != nil
}

func (s *ClientServer) registerGet(w http.ResponseWriter, r *http.Request) {
	utils.RenderTemplate(w, utils.Register, nil)
}

func (s *ClientServer) registerUser(username string, email string, hashedPassword string) bool {
	err := s.DB.AddUser(username, email, hashedPassword)
	if err != nil {
		return false
	}
	middleware.UsersRegistered.Inc()
	return true
}

func (s *ClientServer) registerPost(w http.ResponseWriter, r *http.Request) {
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
		errorMsg = utils.EnterAUsername
	} else if r.FormValue("email") == "" || !strings.Contains(r.FormValue("email"), "@") {
		errorMsg = utils.EnterAValidEmail
	} else if r.FormValue("password") == "" {
		errorMsg = utils.YouHaveToEnterAPassword
	} else if r.FormValue("password") != r.FormValue("password2") {
		errorMsg = utils.PasswordDoesNotMatchMessage
	} else if !s.isUsernameAvailable(r.FormValue("username")) {
		errorMsg = utils.UsernameTaken
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

func (s *ClientServer) registerPostFromJson(w http.ResponseWriter, r *http.Request, registerRequest types.RegisterRequest) {
	latest, latestErr := strconv.ParseInt(r.URL.Query().Get("latest"), 10, 64)
	if latestErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	s.DB.SetLatest(latest)
	error := ""
	if registerRequest.Username == "" {
		error = utils.EnterAUsername
	} else if registerRequest.Email == "" || !strings.Contains(registerRequest.Email, "@") {
		error = utils.EnterAValidEmail
	} else if registerRequest.Pwd == "" {
		error = utils.YouHaveToEnterAPassword
	} else if !s.isUsernameAvailable(registerRequest.Username) {
		error = utils.UsernameTaken
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
