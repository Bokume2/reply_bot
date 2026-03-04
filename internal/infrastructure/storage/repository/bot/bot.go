package bot

import (
	"context"
	"fmt"
	"os"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/interface/schema"

	"github.com/go-ap/activitypub"
	"github.com/go-ap/processing"
)

type BotRepository struct {
	store processing.Store
}

func NewBotRepository(store processing.Store) *BotRepository {
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
