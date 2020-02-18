package main

import (
	"encoding/gob"
	authentication "go/src/authentication"
	"go/src/database"
	"go/src/types"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", authentication.Auth(timeline))
	r.HandleFunc("/public", publicTimeline)
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/addMessage", authentication.Auth(AddMessage)).Methods("POST")
	r.HandleFunc("/login", Login).Methods("GET", "POST")
	r.HandleFunc("/register", Register).Methods("GET", "POST")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	return r
}

func getHTMLTemplate(t *testing.T, resp httptest.ResponseRecorder) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	HTML := string(bodyBytes)
	return HTML
}

func register(username string, password string, password2 string, email string) httptest.ResponseRecorder {
	form := url.Values{}
	if password2 == "" {
		password2 = password
	}
	if email == "" {
		email = username + "@example.com"
	}
	request, _ := http.NewRequest("POST", "/register?username="+username+"&email="+email+"&password="+password+"&password2="+password2, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	return *response
}

func login(username string, password string) httptest.ResponseRecorder {
	form := url.Values{}
	request, _ := http.NewRequest("POST", "/login?username="+username+"&password="+password, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	return *response
}

func register_and_login(username string, password string, password2 string, email string) httptest.ResponseRecorder {
	form := url.Values{}
	request, _ := http.NewRequest("POST", "/register?username="+username+"&email="+email+"&password="+password+"&password2="+password2, strings.NewReader(form.Encode()))
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	request, _ = http.NewRequest("POST", "/login?username="+username+"&password="+password, strings.NewReader(form.Encode()))
	response2 := httptest.NewRecorder()
	Router().ServeHTTP(response2, request)
	return *response2
}

func logout() httptest.ResponseRecorder {
	request, _ := http.NewRequest("GET", "/logout", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	return *response
}

func add_message(text string) httptest.ResponseRecorder {
	request, _ := http.NewRequest("POST", "/add_message?text="+text, nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	return *response
}

func Test_register(t *testing.T) {
	//register(username, email, password, email)
	database.InitDB()

	response := register("user1", "pass1", "pass2", "email")
	assert.Equal(t, 200, response.Code, "Ok response is expected")

	response = register("user1", "pass1", "pass2", "email")
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a username")))

	response = register("user1", "", "", "email")
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a password")))

	response = register("user1", "aa", "bb", "email")
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("The two passwords do not match")))

	response = register("user1", "aa", "aa", "")
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("You have to enter a valid email address")))

}

func Test_login_logout(t *testing.T) {
	database.InitDB()
	gob.Register(&types.User{})
	response := register_and_login("user1", "default", "default", "example@hotmail.com")
	assert.Equal(t, 302, response.Code, "Status found")

	response = logout()
	html := getHTMLTemplate(t, response)
	assert.Equal(t, 302, response.Code, "Status found")

	response = login("user1", "wrongpassword")
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("Invalid password")))

	response = login("user2", "wrongpassword")
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, ("Invalid username")))

}

func Test_message_recording(t *testing.T) {
	database.InitDB()
	gob.Register(&types.User{})
	response := register_and_login("user1", "default", "default", "example@hotmail.com")
	assert.Equal(t, 302, response.Code, "Status found")

	add_message("foo bar 123")
	add_message("hello world 123")

	request, _ := http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	Router().ServeHTTP(&response, request)
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "foo bar 123"))
	assert.True(t, true, strings.Contains(html, "hello world 123"))

}

func Test_timelines(t *testing.T) {
	database.InitDB()
	gob.Register(&types.User{})
	//user1
	response := register_and_login("user1", "default", "default", "example@hotmail.com")
	assert.Equal(t, 302, response.Code, "Status found")

	add_message("the message by user1")
	request, _ := http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	Router().ServeHTTP(&response, request)
	html := getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "the message by user1"))

	response = logout()
	html = getHTMLTemplate(t, response)
	assert.Equal(t, 302, response.Code, "Status found")

	//user2
	response = register_and_login("user2", "default", "default", "example@hotmail.com")
	assert.Equal(t, 302, response.Code, "Status found")

	add_message("the message by user2")
	request, _ = http.NewRequest("GET", "/public", nil)
	response = *httptest.NewRecorder()
	Router().ServeHTTP(&response, request)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "the message by user2"))

	//Follow user1
	request, _ = http.NewRequest("GET", "/user1/follow", nil)
	response = *httptest.NewRecorder()
	Router().ServeHTTP(&response, request)
	html = getHTMLTemplate(t, response)
	assert.True(t, true, strings.Contains(html, "You are currently following this user."))
	assert.Equal(t, 302, response.Code, "Status found")

	//Unfollow user1
	request, _ = http.NewRequest("GET", "/user1/unfollow", nil)
	response = *httptest.NewRecorder()
	Router().ServeHTTP(&response, request)
	html = getHTMLTemplate(t, response)
	assert.Equal(t, 302, response.Code, "Status found")

}

func TestRegisterHandler(t *testing.T) {
	database.InitDB()
	request, _ := http.NewRequest("GET", "/register", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Ok response is expected")
}

func TestLoginHandler(t *testing.T) {
	database.InitDB()
	request, _ := http.NewRequest("GET", "/login", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 200, response.Code, "Ok response is expected")
}

func TestLogoutHandler(t *testing.T) {
	database.InitDB()
	request, _ := http.NewRequest("GET", "/logout", nil)
	response := httptest.NewRecorder()
	Router().ServeHTTP(response, request)
	assert.Equal(t, 302, response.Code, "Ok response is expected")
}
