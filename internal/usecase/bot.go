package usecase

import (
	"context"
	"reply_bot/internal/domain/errors"
	"reply_bot/internal/domain/repository"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
)

type IBotUseCase interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	Reply(ctx context.Context, username string, item *activitypub.Item) (*activitypub.Note, error)
}

type botUseCase struct {
	repo repository.BotRepository
}

func NewBotUseCase(repo repository.BotRepository) IBotUseCase {
	return &botUseCase{repo: repo}
}

func (buc botUseCase) GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error) {
	bot, err := buc.repo.GetByUserName(ctx, username)
	if err != nil {
		if apErrors.IsNotFound(err) {
			return nil, errors.ErrBotNotFound
		}
		return nil, err
	}
	return bot, nil
}

func (buc botUseCase) GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error) {
	bot, err := buc.repo.GetOutBox(ctx, username)
	if err != nil {
		return nil, err
	}
	return bot, nil
}

func (buc botUseCase) Reply(ctx context.Context, username string, item *activitypub.Item) (*activitypub.Note, error) {
	_, err := buc.repo.AddInBox(ctx, username, item)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
