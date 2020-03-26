package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/app/src/client"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/simulator"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

var monitorMiddleware = middleware.Combine(
	middleware.HTTPResponseCodeMonitor,
	middleware.HTTPResponseTimeMonitor,
	middleware.HTTPRequestCountMonitor,
)

func main() {
	gob.Register(&types.User{})
	utils.InitEnvironmentVariables()
	connectionString := "host=127.0.0.1 port=5432 user=" + utils.GetEnvironmentVariable("POSTGRES_USER") + " dbname=" + utils.GetEnvironmentVariable("POSTGRES_DB") + " password=" + utils.GetEnvironmentVariable("POSTGRES_PASSWORD") + " sslmode=disable"
	s := server.CreateNewServer("postgres", connectionString)
	client.AddEndpoints(s, monitorMiddleware)
	simulator.AddEndpoints(s, monitorMiddleware)
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	go utils.ExternalMonitor(utils.GetEnvironmentVariable("EXTERNAL_MONITOR_URL"))
	err := s.Serve(5000)
	if err != nil {
		panic(err)
	}
}
