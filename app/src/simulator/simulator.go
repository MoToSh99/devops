package simulator

import (
	"github.com/matt035343/devops/app/src/client"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
)

//SimulatorServer A wrapping type for Server to handle simulator related endpoints
type SimulatorServer server.Server

//AddEndpoints A wrapper to add simulator endpoints to router
func AddEndpoints(s *server.Server, routeMiddleware middleware.Middleware) {
	sim := SimulatorServer{s.DB, s.Router}
	c := client.ClientServer{DB: s.DB, Router: s.Router}
	r := s.Router.PathPrefix("/simulator").Subrouter()
	r.HandleFunc("/register", routeMiddleware(c.Register)).Methods("GET", "POST")
	r.HandleFunc("/msgs", routeMiddleware(sim.tweetsGet)).Methods("Get")
	r.HandleFunc("/msgs/{username}", routeMiddleware(sim.tweetsUsername)).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", routeMiddleware(sim.followUsername)).Methods("GET", "POST")
	r.HandleFunc("/latest", routeMiddleware(sim.latest)).Methods("GET")
}
