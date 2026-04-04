package usecase

import (
	"context"
	"fmt"
	"math/rand/v2"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Bokume2/reply_bot/internal/domain/errors"
	"github.com/Bokume2/reply_bot/internal/domain/repository"
	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/interface/schema"
	apUtil "github.com/Bokume2/reply_bot/pkg/activitypub"
	htmlUtil "github.com/Bokume2/reply_bot/pkg/html"
	"github.com/Bokume2/reply_bot/pkg/snowflake"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"golang.org/x/text/language"
)

type IBotUseCase interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, username string) (*activitypub.OrderedCollection, error)
	Reply(ctx context.Context, username string, item *activitypub.Item) (*activitypub.Create, *activitypub.Actor, error)
	CancelReply(ctx context.Context, item *activitypub.Item) error
	GetAny(ctx context.Context, id activitypub.IRI) (*activitypub.Item, error)
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
	activity, err := activitypub.ToActivity(*item)
	if err != nil {
		return nil, nil, err
	}
	_, err = buc.repo.AppendToInBox(ctx, username, activity)
	if err != nil {
		return nil, nil, err
	}
	if activity.Type != activitypub.CreateType {
		return nil, nil, nil
	}
	note, err := activitypub.ToObject(activity.Object)
	if err != nil {
		return nil, nil, err
	}
	if note.Type != activitypub.NoteType {
		return nil, nil, nil
	}
	if !(activity.To.Contains(schema.UsernameToID(username)) ||
		activity.Bto.Contains(schema.UsernameToID(username)) ||
		activity.CC.Contains(schema.UsernameToID(username)) ||
		activity.BCC.Contains(schema.UsernameToID(username))) {
		return nil, nil, nil
	}
	content, err := htmlUtil.RemoveHtmlTagsWithRet(note.Content.String())
	if err != nil {
		return nil, nil, err
	}
	mentionRegexp := fmt.Sprintf("@%s(@%s)?", username, config.LocalDomain())
	content = strings.TrimSpace(regexp.MustCompile(mentionRegexp).ReplaceAllString(content, ""))
	replyCont := ""
	for _, v := range config.Dialogues() {
		if content == v.Call {
			replyCont = v.Reply
		}
	}
	var to *activitypub.Actor
	if activity.Actor.IsLink() {
		toItem, err := apUtil.ResolveActivityPubLink(&activity.Actor)
		if err != nil {
			return nil, nil, err
		}
		to, err = activitypub.ToActor(*toItem)
	} else {
		to, err = activitypub.ToActor(activity.Actor)
	}
	if err != nil {
		return nil, nil, err
	}
	if replyCont != "" {
		reply := activitypub.ObjectNew(activitypub.NoteType)
		reply.Content.Set(activitypub.LangRef(language.Japanese), activitypub.Content(replyCont))
		reply.AttributedTo = schema.UsernameToID(username)
		reply.InReplyTo = note.ID
		reply.To.Append(activitypub.PublicNS)
		reply.CC.Append(schema.UsernameToID(username).AddPath(string(activitypub.Followers)))
		reply.CC.Append(activity.Actor.GetID())
		reply.URL = reply.ID
		reply.Published = time.Now()
		noteID := snowflake.TimeToSnowflake(reply.Published, uint16(rand.UintN(0x100)))
		reply.ID = schema.UsernameToID(username).AddPath("/statuses", strconv.FormatUint(noteID, 10))
		_, err = buc.repo.SaveAny(ctx, reply)
		if err != nil {
			return nil, nil, err
		}

		replyAct := activitypub.CreateNew(reply.ID.AddPath("/activity"), reply)
		replyAct.Actor = schema.UsernameToID(username)
		replyAct.Published = reply.Published
		replyAct.To = reply.To
		replyAct.CC = reply.CC
		_, err = buc.repo.SaveAny(ctx, replyAct)
		if err != nil {
			return nil, nil, err
		}

		_, err = buc.repo.AppendToOutBox(ctx, username, replyAct)
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

func (buc botUseCase) GetAny(ctx context.Context, id activitypub.IRI) (*activitypub.Item, error) {
	return buc.repo.LoadAny(ctx, id)
}
