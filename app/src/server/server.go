package server

import (
	"net/http"
	"strconv"

	"github.com/matt035343/devops/app/src/database"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

//Server The HTTP webserver
type Server struct {
	DB     *database.Database
	Router *mux.Router
}

//New Creates a new instance of a server given an instance of a database.
func New(db *database.Database) *Server {
	s := &Server{DB: db}
	s.Router = s.InitRouter()
	return s
}

//Serve Exposes the server on the given port.
func (s *Server) Serve(port int) error {
	return http.ListenAndServe(":"+strconv.Itoa(port), s.Router)
}

//ShutDown Closes and cleans up server, including database
func (s *Server) ShutDown() {
	s.DB.CloseDatabase()
}

//CreateNewServer Creates a new instance of the web server and connects the database
func CreateNewServer(databaseDialect, connectionString string) *Server {
	db, err := database.ConnectDatabase(databaseDialect, connectionString)
	if err != nil {
		panic(err)
	}
	return New(db)
}

//InitRouter Initialises the HTTP routes on the server.
func (s *Server) InitRouter() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/css/").Handler(
		http.StripPrefix("/css/", http.FileServer(http.Dir("src/static/css/"))),
	)

	/* Monitor endpoints */
	r.Handle("/metrics", promhttp.Handler())

	return r
}
