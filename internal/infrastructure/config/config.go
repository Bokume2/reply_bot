package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var (
	LOCAL_DOMAIN           string
	BOT_NAME               string
	BOT_PREFERRED_USERNAME string
	DATA_STORAGE_PATH      string
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env: %v", err)
	}

	LOCAL_DOMAIN = os.Getenv("LOCAL_DOMAIN")
	BOT_NAME = os.Getenv("BOT_NAME")
	BOT_PREFERRED_USERNAME = os.Getenv("BOT_PREFERRED_USERNAME")
	DATA_STORAGE_PATH = os.Getenv("DATA_STORAGE_PATH")
}
