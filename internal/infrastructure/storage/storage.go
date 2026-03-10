package storage

import (
	"log"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/interface/schema"

	apStorage "git.sr.ht/~mariusor/storage-all"
	"github.com/go-ap/activitypub"
	"github.com/go-ap/errors"
	"github.com/go-ap/webfinger"
)

var (
	DataStore        apStorage.FullStorage
	WebFingerStorage webfinger.Storage
)

func InitStorage() {
	var err error
	DataStore, err = apStorage.New(
		apStorage.WithType(apStorage.FS),
		apStorage.WithPath(config.DATA_STORAGE_PATH),
	)
	if err != nil {
		log.Fatal(err)
	}
	err = DataStore.Open()
	if err != nil {
		log.Fatal(err)
	}

	item, err := DataStore.Load(schema.UsernameToID(config.BOT_PREFERRED_USERNAME))
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
