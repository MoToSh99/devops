package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/app/src/client"
	"github.com/matt035343/devops/app/src/log"
	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/simulator"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

var routeMiddleware = middleware.Combine(
	middleware.HTTPResponseCodeMonitor,
	middleware.HTTPResponseTimeMonitor,
	middleware.HTTPRequestCountMonitor,
	middleware.HTTPErrorLoggerMiddleware,
)

func main() {
	log.SetLoggingLevel(log.InfoLevel)
	gob.Register(&types.User{})
	utils.InitEnvironmentVariables()
	connectionString := "host=127.0.0.1 port=5432 user=" + utils.GetEnvironmentVariable("POSTGRES_USER") + " dbname=" + utils.GetEnvironmentVariable("POSTGRES_DB") + " password=" + utils.GetEnvironmentVariable("POSTGRES_PASSWORD") + " sslmode=disable"
	s := server.CreateNewServer("postgres", connectionString)
	client.AddEndpoints(s, routeMiddleware)
	simulator.AddEndpoints(s, routeMiddleware)
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	go utils.ExternalMonitor(utils.GetEnvironmentVariable("EXTERNAL_MONITOR_URL"))
	log.Info("Starting server on port 5000")
	err := s.Serve(5000)
	if err != nil {
		log.CriticalErr("Server stopped with error", err)
		panic(err)
	}
}
