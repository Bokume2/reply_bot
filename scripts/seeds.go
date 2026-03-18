package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage"
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage/repository/bot"

	"github.com/go-ap/activitypub"
)

func CreateBotActor() error {
	actor, err := bot.NewBotRepository(storage.DataStore).CreateBot(context.Background(), config.BOT_PREFERRED_USERNAME, config.BOT_NAME)
	if err != nil {
		return err
	}
	b, _ := activitypub.MarshalJSON(actor)
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("Created actor showed bellow:")
	fmt.Println()
	fmt.Println(buf.String())
	return nil
}

func main() {
	config.LoadEnv()
	storage.InitStorage()

	fmt.Println()
	fmt.Println("Starting to generate seed data...")
	fmt.Println()

	if err := CreateBotActor(); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Finished to initialize data.")
}
