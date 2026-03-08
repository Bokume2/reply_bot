package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/infrastructure/storage"
	"reply_bot/internal/interface/schema"
	"time"

	"github.com/go-ap/activitypub"
	"golang.org/x/text/language"
)

func CreateBotActor() error {
	actor := activitypub.ActorNew(schema.UsernameToId(config.BOT_PREFERRED_USERNAME), activitypub.ServiceType)
	actor.Name.Set(activitypub.LangRef(language.Japanese), activitypub.Content(config.BOT_NAME))
	actor.PreferredUsername.Set(activitypub.LangRef(language.Japanese), activitypub.Content(config.BOT_PREFERRED_USERNAME))
	now := time.Now()
	actor.Published = now
	actor.StartTime = now
	actor.Updated = now
	activitypub.Inbox.AddTo(actor)
	storage.DataStore.Save(activitypub.OrderedCollectionNew(actor.Inbox.GetID()))
	activitypub.Outbox.AddTo(actor)
	storage.DataStore.Save(activitypub.OrderedCollectionNew(actor.Outbox.GetID()))
	activitypub.Following.AddTo(actor)
	storage.DataStore.Save(activitypub.OrderedCollectionNew(actor.Following.GetID()))
	activitypub.Followers.AddTo(actor)
	storage.DataStore.Save(activitypub.OrderedCollectionNew(actor.Followers.GetID()))
	pubkey, err := os.ReadFile(fmt.Sprintf("storage/cred/%s.pub", actor.PreferredUsername.String()))
	if err != nil {
		log.Fatal(err)
	}
	actor.PublicKey = activitypub.PublicKey{
		ID:           activitypub.ID(fmt.Sprintf("%s#main-key", actor.ID.String())),
		Owner:        actor.ID,
		PublicKeyPem: string(pubkey),
	}
	item, err := storage.DataStore.Save(actor)
	if err != nil {
		return err
	}
	b, err := activitypub.MarshalJSON(item)
	var buf bytes.Buffer
	err = json.Indent(&buf, b, "", "  ")
	if err != nil {
		log.Fatal(err)
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
