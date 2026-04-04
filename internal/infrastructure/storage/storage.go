package storage

import (
	"log"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/interface/schema"

	apStorage "git.sr.ht/~mariusor/storage-all"
	"github.com/go-ap/activitypub"
	"github.com/go-ap/errors"
	"github.com/go-ap/webfinger"
)

var (
	DataStore        apStorage.FullStorage
	WebFingerStorage webfinger.Storage
)

func init() {
	var err error
	DataStore, err = apStorage.New(
		apStorage.WithType(apStorage.FS),
		apStorage.WithPath("storage/ap_data"),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = DataStore.Open()
	if err != nil {
		log.Fatal(err)
	}

	item, err := DataStore.Load(schema.UsernameToID(config.BotPreferredUsername()))
	if err != nil {
		if errors.IsNotFound(err) {
			return
		}
		log.Fatal(err)
	}
	actor, err := activitypub.ToActor(item)
	if err != nil {
		log.Fatal(err)
	}
	WebFingerStorage = webfinger.Storage{
		Store: DataStore,
		Root:  *actor,
	}
}
