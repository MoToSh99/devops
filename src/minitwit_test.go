package main

import (
	authentication "go/src/authentication"
	"go/src/database"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", authentication.Auth(timeline))
	r.HandleFunc("/public", publicTimeline)
	r.HandleFunc("/logout", logout)
	r.HandleFunc("/addMessage", authentication.Auth(AddMessage)).Methods("POST")
	r.HandleFunc("/login", Login).Methods("GET", "POST")
	r.HandleFunc("/register", Register).Methods("GET", "POST")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	return r
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

func getHTMLTemplate(t *testing.T, resp httptest.ResponseRecorder) string {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	HTML := string(bodyBytes)
	return HTML

}

// func TestRegister_NoEmail(t *testing.T) {

// 	form := url.Values{}
// 	form.Add("username", "hello")
// 	form.Add("password", "helloa")

// 	database.InitDB()
// 	request, _ := http.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
// 	response := httptest.NewRecorder()
// 	Router().ServeHTTP(response, request)
// 	html := getHTMLTemplate(t, *response)
// 	assert.Equal(t, 200, response.Code, "Ok response is expected")
// 	assert.True(t, true, strings.Contains(html, ("You have to enter a valid email address")))
// }

// func TestRegister_EmptyUsername(t *testing.T) {
// 	register := struct {
// 		username string
// 		password string
// 	}{
// 		"",
// 		"",
// 	}
// 	database.InitDB()
// 	jsonRegister, _ := json.Marshal(register)
// 	request, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(jsonRegister))
// 	response := httptest.NewRecorder()
// 	Router().ServeHTTP(response, request)

// 	html := getHTMLTemplate(t, *response)
// 	assert.Equal(t, 200, response.Code, "Ok response is expected")
// 	assert.True(t, true, strings.Contains(html, ("You have to enter a username")))
// }

// func TestRegisterHandler4(t *testing.T) {
// 	requestBody, err := json.Marshal(map[string]string{
// 		"username":  "magnus",
// 		"email":     "mack@dffsafa.com",
// 		"password":  "",
// 		"password2": "",
// 	})

// 	if err != nil {
// 		fmt.Println(err)
// 	}

// 	database.InitDB()
// 	request, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(requestBody))
// 	response := httptest.NewRecorder()
// 	Router().ServeHTTP(response, request)
// 	html := getHTMLTemplate(t, *response)
// 	assert.Equal(t, 200, response.Code, "Ok response is expected")
// 	assert.Equal(t, true, strings.Contains(html, ("You have to enter a password")))
// }
