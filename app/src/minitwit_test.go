package main

import (
	"encoding/gob"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/types"

	"github.com/stretchr/testify/assert"
)

func getHTMLTemplate(t *testing.T, resp httptest.ResponseRecorder) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	HTML := string(bodyBytes)
	return HTML
}

func register(username string, password string, password2 string, email string, serverInstance *server.Server) httptest.ResponseRecorder {
	form := url.Values{}
	gob.Register(&types.User{})

	if password2 == "" {
		password2 = password
	}
	if email == "" {
		email = username + "@example.com"
	}
	request, _ := http.NewRequest("POST", "/register?username="+username+"&email="+email+"&password="+password+"&password2="+password2, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	return *response
}

func login(username string, password string, serverInstance *server.Server) httptest.ResponseRecorder {
	form := url.Values{}
	request, _ := http.NewRequest("POST", "/login?username="+username+"&password="+password, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	return *response
}

func registerAndLogin(username string, password string, password2 string, email string, serverInstance *server.Server) httptest.ResponseRecorder {
	form := url.Values{}
	gob.Register(&types.User{})
	request, _ := http.NewRequest("POST", "/register?username="+username+"&email="+email+"&password="+password+"&password2="+password2, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	request, _ = http.NewRequest("POST", "/login?username="+username+"&password="+password, strings.NewReader(form.Encode()))
	response2 := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response2, request)
	return *response2
}

func logout(serverInstance *server.Server) httptest.ResponseRecorder {
	request, _ := http.NewRequest("GET", "/logout", nil)
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	return *response
}

func addMessage(text string, serverInstance *server.Server) httptest.ResponseRecorder {
	request, _ := http.NewRequest("POST", "/add_message?text="+text, nil)
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	return *response
}

func initServer() *server.Server {
	connectionString := "host=127.0.0.1 port=5432 user=postgres dbname=minitwit_test password=postgres sslmode=disable"
	clearDatabase(connectionString)
	s := server.CreateNewServer("postgres", connectionString)
	return s
}

func clearDatabase(connectionString string) {
	db, err := gorm.Open("postgres", connectionString)
	if err != nil {
		panic(err)
	}
	db.Delete(&types.User{})
	db.Delete(&types.Follower{})
	db.Delete(&types.Message{})
	db.Delete(&types.LatestResponse{})
}

func TestRegister(t *testing.T) {
	//register(username, email, password, email)
	serverInstance := initServer()

	response := register("user1", "pass1", "pass2", "email", serverInstance)
	assert.Equal(t, 200, response.Code, "Ok response is expected")

	response = register("user1", "pass1", "pass2", "email", serverInstance)
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a username")))

	response = register("user1", "", "", "email", serverInstance)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a password")))

	response = register("user1", "aa", "bb", "email", serverInstance)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("The two passwords do not match")))

	response = register("user1", "aa", "aa", "", serverInstance)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a valid email address")))

	serverInstance.ShutDown()
}

func TestLoginLogout(t *testing.T) {
	serverInstance := initServer()

	response := registerAndLogin("user1", "default", "default", "example@hotmail.com", serverInstance)
	assert.Equal(t, 302, response.Code, "Status found")

	response2 := logout(serverInstance)
	assert.Equal(t, 302, response2.Code, "Status found")

	response3 := login("user1", "wrongpassword", serverInstance)
	html := getHTMLTemplate(t, response3)
	assert.True(t, true, strings.Contains(html, ("Invalid password")))

	response4 := login("user2", "wrongpassword", serverInstance)
	html = getHTMLTemplate(t, response4)
	assert.True(t, true, strings.Contains(html, ("Invalid username")))

	serverInstance.ShutDown()
}

func TestMessageRecording(t *testing.T) {
	serverInstance := initServer()

	gob.Register(&types.User{})
	response := registerAndLogin("user1", "default", "default", "example@hotmail.com", serverInstance)
	assert.Equal(t, 302, response.Code, "Status found")

	addMessage("foo bar 123", serverInstance)
	addMessage("hello world 123", serverInstance)

	request, _ := http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(&response, request)
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "foo bar 123"))
	assert.True(t, true, strings.Contains(html, "hello world 123"))

	serverInstance.ShutDown()
}

func TestTimelines(t *testing.T) {
	serverInstance := initServer()

	gob.Register(&types.User{})
	//user1
	response := registerAndLogin("user1", "default", "default", "example@hotmail.com", serverInstance)
	assert.Equal(t, 302, response.Code, "Status found")

	addMessage("the message by user1", serverInstance)
	request, _ := http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(&response, request)
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "the message by user1"))

	response = logout(serverInstance)
	assert.Equal(t, 302, response.Code, "Status found")

	//user2
	response = registerAndLogin("user2", "default", "default", "example@hotmail.com", serverInstance)
	assert.Equal(t, 302, response.Code, "Status found")

	addMessage("the message by user2", serverInstance)
	request, _ = http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(&response, request)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "the message by user2"))

	//Follow user1
	request, _ = http.NewRequest("GET", "/user1/follow", nil)
	response = *httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(&response, request)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "You are currently following this user."))
	assert.Equal(t, 302, response.Code, "Status found")

	//Unfollow user1
	request, _ = http.NewRequest("GET", "/user1/unfollow", nil)
	response = *httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(&response, request)
	assert.Equal(t, 302, response.Code, "Status found")

	serverInstance.ShutDown()
}

func TestRegisterHandler(t *testing.T) {
	serverInstance := initServer()

	request, _ := http.NewRequest("GET", "/register", nil)
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Ok response is expected")

	serverInstance.ShutDown()
}

func TestLoginHandler(t *testing.T) {
	serverInstance := initServer()

	request, _ := http.NewRequest("GET", "/login", nil)
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Ok response is expected")

	serverInstance.ShutDown()
}

func TestLogoutHandler(t *testing.T) {
	serverInstance := initServer()

	request, _ := http.NewRequest("GET", "/logout", nil)
	response := httptest.NewRecorder()
	serverInstance.Router.ServeHTTP(response, request)
	assert.Equal(t, 302, response.Code, "Ok response is expected")

	serverInstance.ShutDown()
}
