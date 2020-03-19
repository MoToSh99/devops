package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/types"
	"github.com/matt035343/devops/app/src/utils"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	gob.Register(&types.User{})
	utils.InitEnvironmentVariables("../.env")
	connectionString := "host=127.0.0.1 port=5432 user=" + utils.GetEnvironmentVariable("POSTGRES_USER") + " dbname=" + utils.GetEnvironmentVariable("POSTGRES_DB") + " password=" + utils.GetEnvironmentVariable("POSTGRES_PASSWORD") + " sslmode=disable"
	s := server.CreateNewServer("postgres", connectionString)
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	s.Serve(5000)
}
