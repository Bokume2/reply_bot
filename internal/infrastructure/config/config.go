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
	localDomain          string
	localOrigin          activitypub.IRI
	botName              string
	botPreferredUsername string
)

const OriginScheme = "https"

type ReplyDialogue struct {
	Call  string `yaml:"call"`
	Reply string `yaml:"reply"`
}

var dialogues []ReplyDialogue

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env: %v", err)
	}

	localDomain = os.Getenv("LOCAL_DOMAIN")
	localOrigin = activitypub.IRI(fmt.Sprintf("%s://%s", OriginScheme, localDomain))
	botName = os.Getenv("BOT_NAME")
	botPreferredUsername = os.Getenv("BOT_PREFERRED_USERNAME")

	dialoguesFilePath := "reply_dialogues.yaml"
	rawDialogues, err := os.ReadFile(dialoguesFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err = yaml.Unmarshal(rawDialogues, &dialogues); err != nil {
		log.Fatal(err)
	}
}

func LocalDomain() string {
	return localDomain
}

func LocalOrigin() activitypub.IRI {
	return localOrigin
}

func BotName() string {
	return botName
}

func BotPreferredUsername() string {
	return botPreferredUsername
}

func Dialogues() []ReplyDialogue {
	return dialogues
}
