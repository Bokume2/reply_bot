package repository

import (
	"context"

	"github.com/go-ap/activitypub"
)

type BotRepository interface {
	CreateBot(ctx context.Context, username, name string) (*activitypub.Actor, error)
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	AppendToInBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error)
	AppendToOutBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error)
	DeleteFromOutBox(ctx context.Context, item *activitypub.Item) error
	LoadAny(ctx context.Context, id activitypub.IRI) (*activitypub.Item, error)
	SaveAny(ctx context.Context, item activitypub.Item) (*activitypub.Item, error)
}
