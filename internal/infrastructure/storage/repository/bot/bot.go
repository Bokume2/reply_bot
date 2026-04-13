package bot

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/interface/schema"
	"github.com/Bokume2/reply_bot/pkg/sig_key"

	apStorage "git.sr.ht/~mariusor/storage-all"
	"github.com/go-ap/activitypub"
	"github.com/go-ap/errors"
	"golang.org/x/text/language"
)

type BotRepository struct {
	store apStorage.FullStorage
}

func NewBotRepository(store apStorage.FullStorage) *BotRepository {
	return &BotRepository{
		store: store,
	}
}

func (repo BotRepository) CreateBot(ctx context.Context, username, name string) (*activitypub.Actor, error) {
	actor := activitypub.ActorNew(schema.UsernameToID(username), activitypub.ServiceType)
	actor.Name.Set(activitypub.LangRef(language.Japanese), activitypub.Content(name))
	actor.PreferredUsername.Set(activitypub.LangRef(language.Japanese), activitypub.Content(username))
	now := time.Now()
	actor.Published = now
	actor.StartTime = now
	actor.Updated = now
	activitypub.Inbox.AddTo(actor)
	/* Inboxの実体は保存しない(エンドポイントとしてのみ用意) */
	activitypub.Outbox.AddTo(actor)
	_, err := repo.store.Save(activitypub.OrderedCollectionNew(actor.Outbox.GetID()))
	if err != nil {
		return nil, err
	}
	activitypub.Following.AddTo(actor)
	_, err = repo.store.Save(activitypub.OrderedCollectionNew(actor.Following.GetID()))
	if err != nil {
		return nil, err
	}
	activitypub.Followers.AddTo(actor)
	_, err = repo.store.Save(activitypub.OrderedCollectionNew(actor.Followers.GetID()))
	if err != nil {
		return nil, err
	}
	_, err = sig_key.GenerateKeys(actor.PreferredUsername.String())
	if err != nil {
		return nil, err
	}
	pubkey, err := os.ReadFile(sig_key.PubKeyPath(actor.PreferredUsername.String()))
	if err != nil {
		return nil, err
	}
	actor.PublicKey = activitypub.PublicKey{
		ID:           activitypub.ID(fmt.Sprintf("%s#main-key", actor.ID.String())),
		Owner:        actor.ID,
		PublicKeyPem: string(pubkey),
	}
	_, err = repo.store.Save(actor)
	return actor, err
}

func (repo BotRepository) GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error) {
	item, err := repo.store.Load(schema.UsernameToID(username))
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

func (repo BotRepository) GetOutBox(ctx context.Context, bot *activitypub.Actor) (*activitypub.OrderedCollection, error) {
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

func (repo BotRepository) AppendToFollowers(ctx context.Context, bot *activitypub.Actor, id activitypub.IRI) (*activitypub.OrderedCollection, error) {
	item, err := repo.store.Load(bot.Followers.GetID())
	if err != nil {
		return nil, err
	}
	followers, err := activitypub.ToOrderedCollection(item)
	if err != nil {
		return nil, err
	}
	err = followers.Append(id)
	if err != nil {
		return nil, err
	}
	_, err = repo.store.Save(followers)
	if err != nil {
		return nil, err
	}
	return followers, nil
}

func (repo BotRepository) DeleteFromFollowers(ctx context.Context, bot *activitypub.Actor, id activitypub.IRI) error {
	item, err := repo.store.Load(bot.Followers.GetID())
	if err != nil {
		return err
	}
	followers, err := activitypub.ToOrderedCollection(item)
	if err != nil {
		return err
	}
	followers.Remove(id)
	_, err = repo.store.Save(followers)
	return err
}

func (repo BotRepository) AppendToOutBox(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.OrderedCollection, error) {
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

func (repo BotRepository) DeleteFromOutBox(ctx context.Context, item activitypub.Item) error {
	_, err := repo.store.Load((item).GetID())
	if errors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return err
	}
	return repo.store.Delete(item)
}

func (repo BotRepository) LoadAny(ctx context.Context, id activitypub.IRI) (activitypub.Item, error) {
	item, err := repo.store.Load(id)
	return item, err
}

func (repo BotRepository) SaveAny(ctx context.Context, item activitypub.Item) (activitypub.Item, error) {
	it, err := repo.store.Save(item)
	return it, err
}

func (repo BotRepository) updateAvatarOfBot(bot *activitypub.Actor) (*activitypub.Actor, error) {
	if !activitypub.IsNil(bot.Icon) {
		return bot, nil
	}
	path := fmt.Sprintf("public/avatars/%s.jpg", bot.PreferredUsername.String())
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return bot, nil
	}
	url := config.LocalOrigin().AddPath(path)
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
