package config

import (
	"fmt"
	"log"
	"os"

	"github.com/go-ap/activitypub"
	yaml "github.com/goccy/go-yaml"
	"github.com/joho/godotenv"
)

var (
	LOCAL_DOMAIN           string
	LOCAL_ORIGIN           activitypub.IRI
	BOT_NAME               string
	BOT_PREFERRED_USERNAME string
	DATA_STORAGE_PATH      string
)

var Dialogues []ReplyDialogue

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Printf("failed to load .env: %v", err)
	}

	originScheme := "https"

	LOCAL_DOMAIN = os.Getenv("LOCAL_DOMAIN")
	LOCAL_ORIGIN = activitypub.IRI(fmt.Sprintf("%s://%s", originScheme, LOCAL_DOMAIN))
	BOT_NAME = os.Getenv("BOT_NAME")
	BOT_PREFERRED_USERNAME = os.Getenv("BOT_PREFERRED_USERNAME")
	DATA_STORAGE_PATH = os.Getenv("DATA_STORAGE_PATH")

	dialoguesFilePath := "reply_dialogues.yaml"
	rawDialogues, err := os.ReadFile(dialoguesFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(rawDialogues, &Dialogues); err != nil {
		log.Fatal(err)
	}
}

type ReplyDialogue struct {
	Call  string `yaml:"call"`
	Reply string `yaml:"reply"`
}
