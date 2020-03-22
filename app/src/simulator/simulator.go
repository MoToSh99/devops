package simulator

import (
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
)

//Controller A wrapping type for Server to handle simulator related endpoints
type Controller server.Server

//AddEndpoints A wrapper to add simulator endpoints to router
func AddEndpoints(s *server.Server, routeMiddleware middleware.Middleware) {
	sim := Controller{s.DB, s.Router}
	r := s.Router.PathPrefix("/simulator").Subrouter()
	r.HandleFunc("/register", routeMiddleware(sim.register)).Methods("POST")
	r.HandleFunc("/msgs", routeMiddleware(sim.tweetsGet)).Methods("Get")
	r.HandleFunc("/msgs/{username}", routeMiddleware(sim.tweetsUsername)).Methods("GET", "POST")
	r.HandleFunc("/fllws/{username}", routeMiddleware(sim.followUsername)).Methods("GET", "POST")
	r.HandleFunc("/latest", routeMiddleware(sim.latest)).Methods("GET")
}
