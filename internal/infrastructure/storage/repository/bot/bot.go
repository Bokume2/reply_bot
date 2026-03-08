package bot

import (
	"context"
	"fmt"
	"os"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/interface/schema"

	apStorage "git.sr.ht/~mariusor/storage-all"
	"github.com/go-ap/activitypub"
	"github.com/go-ap/errors"
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
	item, err := repo.store.Load(schema.UsernameToId(username))
	if err != nil {
		return nil, err
	}
	bot, err := activitypub.ToActor(item)
	if err != nil {
		return nil, err
	}
	bot, err = repo.updateAvatarOfBot(bot)
	if err != nil {
		return nil, err
	}
	return bot, nil
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
	outbox, err := activitypub.ToOrderedCollection(item)
	if err != nil {
		return nil, err
	}
	return outbox, nil
}

func (repo BotRepository) AppendToInBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error) {
	bot, err := repo.GetByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	item, err := repo.store.Load(bot.Inbox.GetID())
	if err != nil {
		return nil, err
	}
	inbox, err := activitypub.ToOrderedCollection(item)
	if err != nil {
		return nil, err
	}
	err = inbox.Append(activity)
	if err != nil {
		return nil, err
	}
	_, err = repo.store.Save(inbox)
	if err != nil {
		return nil, err
	}
	return inbox, nil
}

func (repo BotRepository) AppendToOutBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error) {
	bot, err := repo.GetByUserName(ctx, username)
	if err != nil {
		return nil, err
	}
	item, err := repo.store.Load(bot.Outbox.GetID())
	if err != nil {
		return nil, err
	}
	outbox, err := activitypub.ToOrderedCollection(item)
	if err != nil {
		return nil, err
	}
	err = outbox.Append(activity)
	if err != nil {
		return nil, err
	}
	_, err = repo.store.Save(outbox)
	if err != nil {
		return nil, err
	}
	return outbox, nil
}

func (repo BotRepository) DeleteFromOutBox(ctx context.Context, item *activitypub.Item) error {
	_, err := repo.store.Load((*item).GetID())
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	return repo.store.Delete(*item)
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
	_, err = repo.store.Save(bot)
	if err != nil {
		return nil, err
	}
	return bot, nil
}
