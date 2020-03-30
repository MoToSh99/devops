package utils

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/matt035343/devops/app/src/log"
)

// InitEnvironmentVariables loads .env file given path
func InitEnvironmentVariables() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	parent := filepath.Dir(wd)
	parent = filepath.Dir(parent)
	err = godotenv.Load(parent + "/.env")
	log.DebugErr("Could not load environment variables", err)
}

// GetEnvironmentVariable returns env var given key
func GetEnvironmentVariable(name string) string {
	return os.Getenv(name)
}
