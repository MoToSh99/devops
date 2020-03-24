package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/app/src/client"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/simulator"
	"github.com/matt035343/devops/app/src/types"

	_ "github.com/mattn/go-sqlite3"
)

var monitorMiddleware = middleware.Combine(
	middleware.HTTPResponseCodeMonitor,
	middleware.HTTPResponseTimeMonitor,
	middleware.HTTPRequestCountMonitor,
)

func main() {
	gob.Register(&types.User{})
	s := server.CreateNewServer("sqlite3", "/tmp/minitwit.db")
	client.AddEndpoints(s, monitorMiddleware)
	simulator.AddEndpoints(s, monitorMiddleware)
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	err := s.Serve(5000)
	if err != nil {
		panic(err)
	}
}
