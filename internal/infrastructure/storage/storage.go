package storage

import (
	"log"
	"reply_bot/internal/infrastructure/config"

	apStorage "git.sr.ht/~mariusor/storage-all"
)

var (
	DataStore apStorage.FullStorage
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
}
