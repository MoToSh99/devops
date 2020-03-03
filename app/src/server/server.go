package server

import (
	"net/http"
	"strconv"

	authentication "github.com/matt035343/devops/app/src/authentication"
	"github.com/matt035343/devops/app/src/database"

	"github.com/gorilla/mux"
)

type Server struct {
	db     *database.Database
	Router *mux.Router
}

func New(db *database.Database) *Server {
	s := &Server{db: db}
	s.Router = s.InitRouter()
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

func (s *Server) InitRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)

	/* Client endpoints */
	r.HandleFunc("/", authentication.Auth(s.timeline))
	r.HandleFunc("/public", s.publicTimeline)
	r.HandleFunc("/logout", s.logout)
	r.HandleFunc("/addMessage", authentication.Auth(s.addMessage)).Methods("POST")
	r.HandleFunc("/login", s.login).Methods("GET", "POST")
	r.HandleFunc("/register", s.register).Methods("GET", "POST")
	r.HandleFunc("/latest", latest).Methods("GET")
	r.HandleFunc("/{username}", authentication.Auth(s.userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(s.followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(s.unfollowUser))

	/* Simulator endpoints */
	r.HandleFunc("/simulator/register", s.register).Methods("GET", "POST")
	r.HandleFunc("/simulator/msgs", s.tweetsGet).Methods("GET")
	r.HandleFunc("/simulator/msgs/{username}", s.tweetsUsername).Methods("GET", "POST")
	r.HandleFunc("/simulator/fllws/{username}", s.followUsername).Methods("GET", "POST")
	return r
}
