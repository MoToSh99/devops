package server

import (
	"net/http"
	"strconv"

	authentication "github.com/matt035343/devops/src/authentication"
	"github.com/matt035343/devops/src/database"

	"github.com/gorilla/mux"
)

type Server struct {
	db     *database.Database
	Router *mux.Router
}

func New(db *database.Database) *Server {
	s := &Server{db: db}
	s.Router = s.initRouter()
	return s
}

func (s *Server) Serve(port int) {
	http.ListenAndServe(":"+strconv.Itoa(port), s.Router)
}

func (s *Server) ShutDown() {
	s.db.CloseDatabase()
}

func CreateNewServer(databaseDialect, connectionString string) *Server {
	db, err := database.ConnectDatabase(databaseDialect, connectionString)
	if err != nil {
		panic(err)
	}
	return New(db)
}

func (s *Server) initRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)
	r.HandleFunc("/", authentication.Auth(s.timeline))
	r.HandleFunc("/public", s.publicTimeline)
	r.HandleFunc("/logout", s.Logout)
	r.HandleFunc("/addMessage", authentication.Auth(s.AddMessage)).Methods("POST")
	r.HandleFunc("/login", s.Login).Methods("GET", "POST")
	r.HandleFunc("/register", s.Register).Methods("GET", "POST")
	r.HandleFunc("/msgs", s.tweetsGet).Methods("Get")
	r.HandleFunc("/msgs/{username}", s.tweetsUsername).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", s.followUsername).Methods("GET", "POST")
	r.HandleFunc("/latest", latest).Methods("GET")
	r.HandleFunc("/{username}", authentication.Auth(s.userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(s.followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(s.unfollowUser))
	return r
}