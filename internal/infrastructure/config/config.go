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
)

const OriginScheme = "https"

type ReplyDialogue struct {
	Call  string `yaml:"call"`
	Reply string `yaml:"reply"`
}

var Dialogues []ReplyDialogue

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	LOCAL_DOMAIN = os.Getenv("LOCAL_DOMAIN")
	LOCAL_ORIGIN = activitypub.IRI(fmt.Sprintf("%s://%s", OriginScheme, LOCAL_DOMAIN))
	BOT_NAME = os.Getenv("BOT_NAME")
	BOT_PREFERRED_USERNAME = os.Getenv("BOT_PREFERRED_USERNAME")

	dialoguesFilePath := "reply_dialogues.yaml"
	rawDialogues, err := os.ReadFile(dialoguesFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(rawDialogues, &Dialogues); err != nil {
		log.Fatal(err)
	}
}
