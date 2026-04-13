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
	htmlUtil "github.com/Bokume2/reply_bot/pkg/html"
	"github.com/Bokume2/reply_bot/pkg/snowflake"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"golang.org/x/text/language"
)

type IBotUseCase interface {
	GetByUserName(ctx context.Context, username string) (*activitypub.Actor, error)
	GetOutBox(ctx context.Context, bot *activitypub.Actor) (*activitypub.OrderedCollection, error)
	AcceptFollowing(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.Accept, error)
	Unfollow(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) error
	Reply(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.Create, error)
	CancelReply(ctx context.Context, activity *activitypub.Activity) error
	GetAny(ctx context.Context, id activitypub.IRI) (activitypub.Item, error)
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

func (buc botUseCase) GetOutBox(ctx context.Context, bot *activitypub.Actor) (*activitypub.OrderedCollection, error) {
	outbox, err := buc.repo.GetOutBox(ctx, bot)
	if err != nil {
		return nil, err
	}
	return outbox, nil
}

func (buc botUseCase) AcceptFollowing(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.Accept, error) {
	actor, _ := activitypub.ToActor(activity.Actor)
	_, err := buc.repo.AppendToFollowers(ctx, bot, actor.GetID())
	if err != nil {
		return nil, err
	}
	accept := activitypub.AcceptNew(activitypub.EmptyID, activity)
	accept.Actor = bot.ID
	accept.Published = time.Now()
	return accept, nil
}

func (buc botUseCase) Unfollow(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) error {
	follow, _ := activitypub.ToActivity(activity.Object)
	return buc.repo.DeleteFromFollowers(ctx, bot, follow.Actor.GetID())
}

func (buc botUseCase) Reply(ctx context.Context, bot *activitypub.Actor, activity *activitypub.Activity) (*activitypub.Create, error) {
	note, _ := activitypub.ToObject(activity.Object)
	content, err := htmlUtil.RemoveHtmlTagsWithRet(note.Content.String())
	if err != nil {
		return nil, err
	}
	mentionRegexp := fmt.Sprintf("@%s(@%s)?", bot.PreferredUsername, config.LocalDomain())
	content = strings.TrimSpace(regexp.MustCompile(mentionRegexp).ReplaceAllString(content, ""))
	replyCont := ""
	for _, v := range config.Dialogues() {
		if content == v.Call {
			replyCont = v.Reply
		}
	}
	if replyCont != "" {
		reply := activitypub.ObjectNew(activitypub.NoteType)
		reply.Content.Set(activitypub.LangRef(language.Japanese), activitypub.Content(replyCont))
		reply.AttributedTo = bot.ID
		reply.InReplyTo = note.ID
		reply.To.Append(activitypub.PublicNS)
		reply.CC.Append(bot.ID.AddPath(string(activitypub.Followers)))
		reply.CC.Append(activity.Actor.GetID())
		reply.URL = reply.ID
		reply.Published = time.Now()
		noteID := snowflake.TimeToSnowflake(reply.Published, uint16(rand.UintN(0x100)))
		reply.ID = bot.ID.AddPath("/statuses", strconv.FormatUint(noteID, 10))
		_, err = buc.repo.SaveAny(ctx, reply)
		if err != nil {
			return nil, err
		}

		replyAct := activitypub.CreateNew(reply.ID.AddPath("/activity"), reply)
		replyAct.Actor = bot.ID
		replyAct.Published = reply.Published
		replyAct.To = reply.To
		replyAct.CC = reply.CC
		_, err = buc.repo.SaveAny(ctx, replyAct)
		if err != nil {
			return nil, err
		}

		_, err = buc.repo.AppendToOutBox(ctx, bot, replyAct)
		if err != nil {
			return nil, err
		}
		return replyAct, nil
	}
	return nil, nil
}

func (buc botUseCase) CancelReply(ctx context.Context, activity *activitypub.Activity) error {
	return buc.repo.DeleteFromOutBox(ctx, activity)
}

func (buc botUseCase) GetAny(ctx context.Context, id activitypub.IRI) (activitypub.Item, error) {
	return buc.repo.LoadAny(ctx, id)
}
