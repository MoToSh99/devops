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

<<<<<<< HEAD
var monitorMiddleware = middleware.Combine(
	middleware.HTTPResponseCodeMonitor,
	middleware.HTTPResponseTimeMonitor,
	middleware.HTTPRequestCountMonitor,
)

func (s *Server) initRouter() *mux.Router {
=======
func (s *Server) InitRouter() *mux.Router {
>>>>>>> develop
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)

	/* Monitor endpoints */
	r.Handle("/metrics", promhttp.Handler())

	/* Client endpoints */
<<<<<<< HEAD
	r.HandleFunc("/", middleware.Auth(monitorMiddleware(s.timeline)))
	r.HandleFunc("/public", monitorMiddleware(s.publicTimeline))
	r.HandleFunc("/logout", monitorMiddleware(s.Logout))
	r.HandleFunc("/addMessage", monitorMiddleware(middleware.Auth(s.AddMessage))).Methods("POST")
	r.HandleFunc("/login", monitorMiddleware(s.Login)).Methods("GET", "POST")
	r.HandleFunc("/register", monitorMiddleware(s.Register)).Methods("GET", "POST")
	r.HandleFunc("/{username}", monitorMiddleware(middleware.Auth(s.userTimeline)))
	r.HandleFunc("/{username}/follow", monitorMiddleware(middleware.Auth(s.followUser)))
	r.HandleFunc("/{username}/unfollow", monitorMiddleware(middleware.Auth(s.unfollowUser)))

	/* Simulator endpoints */
	r.HandleFunc("/simulator/register", monitorMiddleware(s.Register)).Methods("GET", "POST")
	r.HandleFunc("/simulator/msgs", monitorMiddleware(s.tweetsGet)).Methods("Get")
	r.HandleFunc("/simulator/msgs/{username}", monitorMiddleware(s.tweetsUsername)).Methods("GET", "POST")
	r.HandleFunc("/simulator/fllws/{username}", monitorMiddleware(s.followUsername)).Methods("GET", "POST")
	r.HandleFunc("/simulator/latest", monitorMiddleware(s.latest)).Methods("GET")

=======
	r.HandleFunc("/", authentication.Auth(s.timeline))
	r.HandleFunc("/public", s.publicTimeline)
	r.HandleFunc("/logout", s.logout)
	r.HandleFunc("/addMessage", authentication.Auth(s.addMessage)).Methods("POST")
	r.HandleFunc("/login", s.login).Methods("GET", "POST")
	r.HandleFunc("/register", s.register).Methods("GET", "POST")
	r.HandleFunc("/{username}", authentication.Auth(s.userTimeline))
	r.HandleFunc("/{username}/follow", authentication.Auth(s.followUser))
	r.HandleFunc("/{username}/unfollow", authentication.Auth(s.unfollowUser))

	/* Simulator endpoints */
	r.HandleFunc("/simulator/register", s.register).Methods("GET", "POST")
	r.HandleFunc("/simulator/msgs", s.tweetsGet).Methods("GET")
	r.HandleFunc("/simulator/msgs/{username}", s.tweetsUsername).Methods("GET", "POST")
	r.HandleFunc("/simulator/fllws/{username}", s.followUsername).Methods("GET", "POST")
	r.HandleFunc("/simulator/latest", s.latest).Methods("GET")
>>>>>>> develop
	return r
}
