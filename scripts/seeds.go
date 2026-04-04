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
	"github.com/Bokume2/reply_bot/internal/interface/schema"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
)

func CreateBotActor() error {
	// 既にActorが存在する場合は何もしない
	it, err := storage.DataStore.Load(schema.UsernameToID(config.BotPreferredUsername()))
	if err != nil && !apErrors.IsNotFound(err) {
		return err
	}
	actor, err := activitypub.ToActor(it)
	if actor != nil && err == nil {
		return nil
	}

	actor, err = bot.NewBotRepository(storage.DataStore).CreateBot(context.Background(), config.BotPreferredUsername(), config.BotName())
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
	fmt.Println()
	fmt.Println("Starting to generate seed data...")
	fmt.Println()

	if err := CreateBotActor(); err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("Finished to initialize data.")
}
