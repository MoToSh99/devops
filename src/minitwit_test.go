package main

import (
	authentication "go/src/authentication"
	"go/src/database"
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
