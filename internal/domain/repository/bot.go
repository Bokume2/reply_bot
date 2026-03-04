package repository

import (
	"context"

	"github.com/go-ap/activitypub"
)

type BotRepository interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
}
