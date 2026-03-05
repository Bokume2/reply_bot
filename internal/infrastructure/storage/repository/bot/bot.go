package bot

import (
	"context"
	"fmt"
	"os"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/interface/schema"

	apStorage "git.sr.ht/~mariusor/storage-all"
	"github.com/go-ap/activitypub"
)

type BotRepository struct {
	store apStorage.FullStorage
}

func NewBotRepository(store apStorage.FullStorage) *BotRepository {
	return &BotRepository{
		store: store,
	}
}

func (repo BotRepository) GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error) {
	bot, err := repo.store.Load(schema.UsernameToId(username))
	if err != nil {
		return nil, err
	}
	var botActor *activitypub.Actor
	err = activitypub.OnActor(bot, func(a *activitypub.Actor) error {
		var innerErr error
		botActor, innerErr = repo.updateAvatarOfBot(a)
		return innerErr
	})
	if err != nil {
		return nil, err
	}
	return botActor, nil
}

func (repo BotRepository) GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error) {
	bot, err := repo.GetByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	item, err := repo.store.Load(bot.Outbox.GetID())
	if err != nil {
		return nil, err
	}
	var outBox *activitypub.OrderedCollection
	err = activitypub.OnOrderedCollection(item, func(oc *activitypub.OrderedCollection) error {
		outBox = oc
		return nil
	})
	if err != nil {
		return nil, err
	}
	return outBox, nil
}

func (repo BotRepository) updateAvatarOfBot(bot *activitypub.Actor) (*activitypub.Actor, error) {
	if !activitypub.IsNil(bot.Image) {
		return bot, nil
	}
	path := fmt.Sprintf("public/avatars/%s.jpg", bot.PreferredUsername.String())
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return bot, nil
	}
	url := config.LOCAL_ORIGIN.AddPath(path)
	avatar := activitypub.Image{
		Type:      activitypub.ImageType,
		MediaType: activitypub.MimeType("image/jpeg"),
		URL:       url,
	}
	bot.Icon = avatar
	var botItem activitypub.Item
	botItem, err = repo.store.Save(bot)
	err = activitypub.OnActor(botItem, func(a *activitypub.Actor) error {
		bot = a
		return nil
	})
	if err != nil {
		return nil, err
	}
	return bot, nil
}
