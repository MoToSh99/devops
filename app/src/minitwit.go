package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/app/src/middleware"
	"github.com/matt035343/devops/app/src/server"
	"github.com/matt035343/devops/app/src/types"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	gob.Register(&types.User{})
	s := server.CreateNewServer("sqlite3", "/tmp/minitwit.db")
	defer s.ShutDown()
	go middleware.HTTPRequestCounter()
	err := s.Serve(5000)
	if err != nil {
		panic(err)
	}
}
