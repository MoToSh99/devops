package utils

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// InitEnvironmentVariables loads .env file given path
func InitEnvironmentVariables(path string) {
	err := godotenv.Load(path)
	if err != nil {
		fmt.Println("No env file loaded")
	}
}

// GetEnvironmentVariable returns env var given key
func GetEnvironmentVariable(name string) string {
	return os.Getenv(name)
}
