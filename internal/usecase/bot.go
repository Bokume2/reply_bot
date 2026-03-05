package usecase

import (
	"context"
	"fmt"
	"regexp"
	"reply_bot/internal/domain/errors"
	"reply_bot/internal/domain/repository"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/interface/schema"
	"reply_bot/internal/utils"
	"strings"
	"time"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"golang.org/x/text/language"
)

type IBotUseCase interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	Reply(ctx context.Context, username string, item *activitypub.Item) (*activitypub.Create, *activitypub.Actor, error)
	CancelReply(ctx context.Context, item *activitypub.Item) error
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

func (buc botUseCase) Reply(ctx context.Context, username string, item *activitypub.Item) (*activitypub.Create, *activitypub.Actor, error) {
	var activity *activitypub.Activity
	err := activitypub.OnActivity(*item, func(a *activitypub.Activity) error {
		activity = a
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	var to *activitypub.Actor
	err = activitypub.OnActor(activity.Actor, func(a *activitypub.Actor) error {
		to = a
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	_, err = buc.repo.AddInBox(ctx, username, activity)
	if err != nil {
		return nil, nil, err
	}
	var note *activitypub.Note
	err = activitypub.OnObject(activity.Object, func(o *activitypub.Object) error {
		note = o
		return nil
	})
	content, err := utils.RemoveHtmlTagsWithRet(note.Content.String())
	if err != nil {
		return nil, nil, err
	}
	mentionRegexp := fmt.Sprintf("@%s(@%s)?", username, config.LOCAL_DOMAIN)
	content = strings.TrimSpace(regexp.MustCompile(mentionRegexp).ReplaceAllString(content, ""))
	replyCont := ""
	for _, v := range config.Dialogues {
		if content == v.Call {
			replyCont = v.Reply
		}
	}
	if replyCont != "" {
		reply := activitypub.ObjectNew(activitypub.NoteType)
		reply.Content.Set(activitypub.LangRef(language.Japanese), activitypub.Content(replyCont))
		reply.AttributedTo = schema.UsernameToId(username)
		reply.InReplyTo = activity.ID
		reply.To.Append(activitypub.PublicNS)
		reply.CC.Append(schema.UsernameToId(username).AddPath(string(activitypub.Followers)))
		reply.CC.Append(activity.Actor.GetID())
		reply.URL = reply.ID
		reply.Published = time.Now()

		replyAct := activitypub.CreateNew(activitypub.EmptyID, reply)
		replyAct.Actor = schema.UsernameToId(username)
		replyAct.Published = reply.Published
		replyAct.To = reply.To
		replyAct.CC = reply.CC

		_, err = buc.repo.AddOutBox(ctx, username, replyAct)
		if err != nil {
			return nil, nil, err
		}
		return replyAct, to, nil
	}
	return nil, nil, nil
}

func (buc botUseCase) CancelReply(ctx context.Context, item *activitypub.Item) error {
	return buc.repo.DeleteFromOutBox(ctx, item)
}
