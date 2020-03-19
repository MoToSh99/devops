package client

import (
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
)

//Controller A wrapping type for Server to handle client related endpoints
type Controller server.Server

//AddEndpoints Adds client-related endpoints to Server instance
func AddEndpoints(s *server.Server, routeMiddleware middleware.Middleware) {
	c := &Controller{DB: s.DB, Router: s.Router}
	r := s.Router
	r.HandleFunc("/", routeMiddleware(middleware.Auth(c.timeline)))
	r.HandleFunc("/public", routeMiddleware(c.publicTimeline))
	r.HandleFunc("/logout", routeMiddleware(c.logout))
	r.HandleFunc("/addMessage", routeMiddleware(middleware.Auth(c.addMessage))).Methods("POST")
	r.HandleFunc("/login", routeMiddleware(c.login)).Methods("GET", "POST")
	r.HandleFunc("/register", routeMiddleware(c.register)).Methods("GET", "POST")
	r.HandleFunc("/{username}", routeMiddleware(middleware.Auth(c.userTimeline)))
	r.HandleFunc("/{username}/follow", routeMiddleware(middleware.Auth(c.followUser)))
	r.HandleFunc("/{username}/unfollow", routeMiddleware(middleware.Auth(c.unfollowUser)))
}
