package client

import (
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	authentication "github.com/matt035343/devops/app/src/authentication"
	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"
	"golang.org/x/crypto/bcrypt"
)

var perPage = 30

func (c *Controller) timeline(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user").(*(types.User))

	messages, err := c.DB.GetTimelineViewMessages(user.UserID, perPage)
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
	err = utils.RenderTemplate(w, utils.Timeline, data)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) publicTimeline(w http.ResponseWriter, r *http.Request) {
	messages, err := c.DB.GetPublicViewMessages(perPage)
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

	err = utils.RenderTemplate(w, utils.Timeline, data)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) userTimeline(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	profile, err := c.DB.GetUserFromUsername(username)
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
		err = utils.RenderTemplate(w, utils.Timeline, error_response)
		if err != nil {
			panic(err)
		}
		return
	} else if err != nil {
		panic(err)
	}

	follower, err := c.DB.GetFollower(userID, profile.UserID)
	if err != nil && err != gorm.ErrRecordNotFound {
		panic(err)
	}

	messages, err := c.DB.GetUserViewMessages(profile.UserID, perPage)
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

	err = utils.RenderTemplate(w, utils.Timeline, data)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) followUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := c.DB.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = c.DB.AddFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	middleware.UsersFollowed.Inc()
	authentication.Flash(w, r, "You are now following "+username)

	http.Redirect(w, r, "/"+username, http.StatusFound)
}

func (c *Controller) unfollowUser(w http.ResponseWriter, r *http.Request) {
	username := mux.Vars(r)["username"]
	user, err := c.DB.GetUserFromUsername(username)
	if err != nil {
		http.Redirect(w, r, "/public", http.StatusNotFound)
	}
	sessionUser := authentication.GetSessionValue(w, r, "user")
	sessionUserID := (sessionUser.(*types.User)).UserID
	err = c.DB.DeleteFollower(sessionUserID, user.UserID)
	if err != nil {
		panic(err)
	}
	middleware.UsersUnfollowed.Inc()
	authentication.Flash(w, r, "You are no longer following "+username)

	http.Redirect(w, r, "/", http.StatusFound)
}

func (c *Controller) addMessage(w http.ResponseWriter, r *http.Request) {
	user := authentication.GetSessionValue(w, r, "user")
	text := r.FormValue("text")
	userID := (user.(*types.User)).UserID

	if text != "" {
		err := c.DB.AddMessage(userID, text, time.Now())
		if err != nil {
			panic(err)
		}
		middleware.MessagesSent.Inc()
	}

	http.Redirect(w, r, "/public", http.StatusFound)
}

func (c *Controller) login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.loginGet(w, r)
	} else if r.Method == "POST" {
		c.loginPost(w, r)
	}

}

func (c *Controller) loginGet(w http.ResponseWriter, r *http.Request) {
	err := utils.RenderTemplate(w, utils.Login, nil)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) loginPost(w http.ResponseWriter, r *http.Request) {
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

	userFound, user := c.authenticate(username, password)

	if !userFound {
		data.HasError = true
		data.ErrorMsg = "Invalid password"
		data.IsLoggedIn = false
		err := utils.RenderTemplate(w, utils.Login, data)
		if err != nil {
			panic(err)
		}
		return
	}
	if errorMsg != "" {
		log.Warning("Login failed for username %s with error message: ", username, errorMsg)
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

func (c *Controller) authenticate(username string, password string) (bool, *types.User) {
	user, err := c.DB.GetUserFromUsername(username)
	if err != nil && err != gorm.ErrRecordNotFound {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return false, nil
	}
	return true, &user
}

func (c *Controller) logout(w http.ResponseWriter, r *http.Request) {
	err := authentication.ClearSession(w, r)
	if err != nil {
		panic(err)
	}
	http.Redirect(w, r, "/public", http.StatusFound)

}

func (c *Controller) register(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		c.registerGet(w, r)
	} else if r.Method == "POST" {
		c.registerPost(w, r)
	}

}

func (c *Controller) registerGet(w http.ResponseWriter, r *http.Request) {
	err := utils.RenderTemplate(w, utils.Register, nil)
	if err != nil {
		panic(err)
	}
}

func (c *Controller) registerUser(username string, email string, hashedPassword string) bool {
	err := c.DB.AddUser(username, email, hashedPassword)
	if err != nil {
		return false
	}
	middleware.UsersRegistered.Inc()
	return true
}

func (c *Controller) registerPost(w http.ResponseWriter, r *http.Request) {
	usernameFromForm := r.FormValue("username")
	errorMsg := ""

	if usernameFromForm == "" {
		errorMsg = utils.EnterAUsername
	} else if r.FormValue("email") == "" || !strings.Contains(r.FormValue("email"), "@") {
		errorMsg = utils.EnterAValidEmail
	} else if r.FormValue("password") == "" {
		errorMsg = utils.YouHaveToEnterAPassword
	} else if r.FormValue("password") != r.FormValue("password2") {
		errorMsg = utils.PasswordDoesNotMatchMessage
	} else if len(r.FormValue("password")) < 8 {
		errorMsg = utils.PasswordMustBeAtleast8Chars
	} else if !utils.StringContainsUppercase(r.FormValue("password")) {
		errorMsg = utils.PasswordMustContainAtleastOneUppercase
	} else if !c.DB.IsUsernameAvailable(r.FormValue("username")) {
		errorMsg = utils.UsernameTaken
	} else {
		hashedPasswordInBytes, _ := bcrypt.GenerateFromPassword([]byte(r.FormValue("password")), 14)
		c.registerUser(r.FormValue("username"), r.FormValue("email"), string(hashedPasswordInBytes))
	}
	if errorMsg != "" {
		data := struct {
			HasError   bool
			ErrorMsg   string
			IsLoggedIn bool
		}{true, errorMsg, false}
		err := utils.RenderTemplate(w, utils.Register, data)
		if err != nil {
			panic(err)
		}

	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}
