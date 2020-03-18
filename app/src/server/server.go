package server

import (
	"net/http"
	"strconv"

	"github.com/matt035343/devops/app/src/database"
	"github.com/matt035343/devops/app/src/middleware"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

func (s *Server) Serve(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), s.Router)
}

func (s *Server) ShutDown() error {
	return s.db.CloseDatabase()
}

func CreateNewServer(databaseDialect, connectionString string) *Server {
	db, err := database.ConnectDatabase(databaseDialect, connectionString)
	if err != nil {
		panic(err)
	}
	return New(db)
}

var monitorMiddleware = middleware.Combine(
	middleware.HTTPResponseCodeMonitor,
	middleware.HTTPResponseTimeMonitor,
	middleware.HTTPRequestCountMonitor,
)

func (s *Server) InitRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)

	/* Monitor endpoints */
	r.Handle("/metrics", promhttp.Handler())

	/* Client endpoints */
	r.HandleFunc("/", middleware.Auth(monitorMiddleware(s.timeline)))
	r.HandleFunc("/public", monitorMiddleware(s.publicTimeline))
	r.HandleFunc("/logout", monitorMiddleware(s.logout))
	r.HandleFunc("/addMessage", monitorMiddleware(middleware.Auth(s.addMessage))).Methods("POST")
	r.HandleFunc("/login", monitorMiddleware(s.login)).Methods("GET", "POST")
	r.HandleFunc("/register", monitorMiddleware(s.register)).Methods("GET", "POST")
	r.HandleFunc("/{username}", monitorMiddleware(middleware.Auth(s.userTimeline)))
	r.HandleFunc("/{username}/follow", monitorMiddleware(middleware.Auth(s.followUser)))
	r.HandleFunc("/{username}/unfollow", monitorMiddleware(middleware.Auth(s.unfollowUser)))

	/* Simulator endpoints */
	r.HandleFunc("/simulator/register", monitorMiddleware(s.register)).Methods("GET", "POST")
	r.HandleFunc("/simulator/msgs", monitorMiddleware(s.tweetsGet)).Methods("Get")
	r.HandleFunc("/simulator/msgs/{username}", monitorMiddleware(s.tweetsUsername)).Methods("GET", "POST")
	r.HandleFunc("/simulator/fllws/{username}", monitorMiddleware(s.followUsername)).Methods("GET", "POST")
	r.HandleFunc("/simulator/latest", monitorMiddleware(s.latest)).Methods("GET")

	return r
}
