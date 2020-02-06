package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {

	fmt.Println("Hello Mack")

	r := mux.NewRouter()
	r.HandleFunc("/", timeline)
	r.HandleFunc("/public", public_timeline)
	r.HandleFunc("/{username}", user_timeline)
	r.HandleFunc("/{username}/follow", follow_user)
	r.HandleFunc("/{username}/unfollow", unfollow_user)
	r.HandleFunc("/logout", logout)

	r.HandleFunc("/add_message", add_message).Methods("POST")
	r.HandleFunc("/login", login).Methods("GET", "POST")
	r.HandleFunc("/register", register).Methods("GET", "POST")

	http.ListenAndServe(":5000", r)

}

func timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Timeline hit")
}

func public_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "public_timeline hit")
}

func user_timeline(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "user_timeline hit")
}
func follow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "follow_user hit")
}
func unfollow_user(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "unfollow_user")
}
func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "logout hit")
}
func add_message(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "add_message hit")
}
func login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "login hit")
}
func register(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "register hit")
}
