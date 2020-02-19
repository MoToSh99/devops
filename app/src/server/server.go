package server

import (
	"go/src/database"
	"net/http"

	"github.com/gorilla/mux"
)

func CreateNewServer(databaseDialect, connectionString string, createDatabaseIfAbsent bool, schemaLocation string) *types.Server {
	db, err := database.ConnectDatabase(databaseDialect, connectionString, createDatabaseIfAbsent, schemaLocation)
	if err != nil {
		panic(err)
	}
	return &types.Server{
		router: initRouter(),
		db:     db,
	}
}

func initRouter() *Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)
	r.HandleFunc("/", authentication.Auth(timeline))
	r.HandleFunc("/public", publicTimeline)
	r.HandleFunc("/logout", Logout)
	r.HandleFunc("/addMessage", authentication.Auth(AddMessage)).Methods("POST")
	r.HandleFunc("/login", Login).Methods("GET", "POST")
	r.HandleFunc("/register", Register).Methods("GET", "POST")
	r.HandleFunc("/msgs", tweetsGet).Methods("Get")
	r.HandleFunc("/msgs/{username}", tweetsUsername).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", followUsername).Methods("GET", "POST")
	r.HandleFunc("/latest", latest).Methods("GET")
	r.HandleFunc("/{username}", authentication.Auth(userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(unfollowUser))
	return r
}

func (s *Server) Serve(port int) {
	http.ListenAndServe(":"+port, s.router)
}
