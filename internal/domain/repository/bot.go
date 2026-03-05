package repository

import (
	"context"

	"github.com/go-ap/activitypub"
)

type BotRepository interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	AddInBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error)
	AddOutBox(ctx context.Context, username string, activity *activitypub.Activity) (*activitypub.OrderedCollection, error)
	DeleteFromOutBox(ctx context.Context, item *activitypub.Item) error
}
