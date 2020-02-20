package main

import (
	"encoding/gob"

	"github.com/matt035343/devops/src/server"
	"github.com/matt035343/devops/src/types"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	gob.Register(&types.User{})
	s := server.CreateNewServer("sqlite3", "/tmp/minitwit.db")
	defer s.ShutDown()
	s.Serve(5000)
}
