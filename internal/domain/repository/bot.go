package repository

import (
	"context"

	"github.com/go-ap/activitypub"
)

type BotRepository interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	AddInBox(ctx context.Context, username string, item *activitypub.Item) (*activitypub.OrderedCollection, error)
}
