package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/types"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	gob.Register(&types.User{})
	//s := server.CreateNewServer("sqlite3", "/tmp/minitwit.db")
	connectionString := "host=127.0.0.1 port=5432 user=" + os.Getenv("POSTGRES_USER") + " dbname=" + os.Getenv("POSTGRES_DB") + " password=" + os.Getenv("POSTGRES_PASSWORD") + " sslmode=disable"
	s := server.CreateNewServer("postgres", connectionString)
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	s.Serve(5000)
}
