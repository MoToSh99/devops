package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/matt035343/devops/src/database"

	"github.com/gorilla/mux"
)

type Server struct {
	db     *database.Database
	router *mux.Router
}

func New(db *database.Database, router *mux.Router) *Server {
	return &Server{db: db, router: router}
}

func (s *Server) Serve(port int) {
	http.ListenAndServe(":"+strconv.Itoa(port), s.router)
}

func CreateNewServer(databaseDialect, connectionString string) *Server {
	db, err := database.ConnectDatabase(databaseDialect, connectionString)
	if err != nil {
		panic(err)
	}
	l, err := db.GetFollowers(1, -1)
	fmt.Println(l)
	u, err := db.GetUser(1)
	fmt.Println(u)
	return New(db, initRouter())
}

func initRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)
	// r.HandleFunc("/", authentication.Auth(timeline))
	// r.HandleFunc("/public", publicTimeline)
	// r.HandleFunc("/logout", Logout)
	// r.HandleFunc("/addMessage", authentication.Auth(AddMessage)).Methods("POST")
	// r.HandleFunc("/login", Login).Methods("GET", "POST")
	// r.HandleFunc("/register", Register).Methods("GET", "POST")
	// r.HandleFunc("/msgs", tweetsGet).Methods("Get")
	// r.HandleFunc("/msgs/{username}", tweetsUsername).Methods("GET", "POST")
	// r.HandleFunc("/fllws/{username}", followUsername).Methods("GET", "POST")
	// r.HandleFunc("/latest", latest).Methods("GET")
	// r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	// r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	// r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	return r
}
