package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
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
	if err != nil {
		fmt.Println(err)
	}
}

// GetEnvironmentVariable returns env var given key
func GetEnvironmentVariable(name string) string {
	return os.Getenv(name)
}
