package repository

import (
	"context"

	"github.com/go-ap/activitypub"
)

type BotRepository interface {
	CreateBot(ctx context.Context, username, name string) (*activitypub.Actor, error)
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, bot *activitypub.Actor) (*activitypub.OrderedCollection, error)
	AppendToFollowers(ctx context.Context, bot *activitypub.Actor, id activitypub.IRI) (*activitypub.OrderedCollection, error)
	DeleteFromFollowers(ctx context.Context, bot *activitypub.Actor, id activitypub.IRI) error
	AppendToOutBox(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.OrderedCollection, error)
	DeleteFromOutBox(ctx context.Context, item activitypub.Item) error
	LoadAny(ctx context.Context, id activitypub.IRI) (activitypub.Item, error)
	SaveAny(ctx context.Context, item activitypub.Item) (activitypub.Item, error)
}
