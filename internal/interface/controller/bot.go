package controller

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	domainErrors "github.com/Bokume2/reply_bot/internal/domain/errors"
	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
	"github.com/Bokume2/reply_bot/internal/interface/schema"
	"github.com/Bokume2/reply_bot/internal/usecase"
	apUtil "github.com/Bokume2/reply_bot/pkg/activitypub"
	jldUtil "github.com/Bokume2/reply_bot/pkg/jsonld"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"github.com/labstack/echo/v5"
)

type BotController struct {
	buc usecase.IBotUseCase
}

func NewBotController(buc usecase.IBotUseCase) *BotController {
	return &BotController{
		buc: buc,
	}
}

func BotsMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		err := next(c)
		if errors.Is(err, domainErrors.ErrBotNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return err
	}
}

func (bc BotController) GetByUserName(c *echo.Context) error {
	bot, err := bc.buc.GetByUserName(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	return jldUtil.JSONLDResponse(c, http.StatusOK, bot)
}

func (bc BotController) GetOutBox(c *echo.Context) error {
	bot, err := bc.buc.GetByUserName(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	outbox, err := bc.buc.GetOutBox(c.Request().Context(), bot)
	if err != nil {
		return err
	}
	return jldUtil.JSONLDResponse(c, http.StatusOK, outbox)
}

func (bc BotController) PostInBox(c *echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) != "application/activity+json" && c.Request().Header.Get(echo.HeaderContentType) != "application/ld+json; profile=\"https://www.w3.org/ns/activitystreams\"" {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "expected application/activity+json")
	}
	ab := ActivityBinder{}
	postedActivity := new(activitypub.Activity)
	if err := ab.Bind(c, postedActivity); err != nil {
		return err
	}
	it, err := apUtil.ResolveActivityPubLink(postedActivity.Actor)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Activity must have valid actor information")
	}
	postedActor, err := activitypub.ToActor(it)
	if err != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Activity must have valid actor information")
	}
	postedActivity.Actor = postedActor
	if err = bc.verifyRequest(c.Request(), postedActor); err != nil {
		return err
	}
	bot, err := bc.buc.GetByUserName(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	switch postedActivity.Type {
	case activitypub.CreateType:
		it, err := apUtil.ResolveActivityPubLink(postedActivity.Object)
		if err != nil {
			return err
		}
		note, err := activitypub.ToObject(it)
		if err != nil {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "object of activity must be Object or Linke type")
		}
		postedActivity.Object = note
		if note.Type != activitypub.NoteType {
			return c.NoContent(http.StatusAccepted)
		}
		if !(postedActivity.To.Contains(bot.ID) ||
			postedActivity.Bto.Contains(bot.ID) ||
			postedActivity.CC.Contains(bot.ID) ||
			postedActivity.BCC.Contains(bot.ID)) {
			return c.NoContent(http.StatusAccepted)
		}
		return bc.handleReply(c, bot, postedActivity)
	case activitypub.FollowType:
		if !postedActivity.Object.GetID().Equal(bot.ID) {
			return c.NoContent(http.StatusAccepted)
		}
		return bc.handleFollow(c, bot, postedActivity)
	case activitypub.UndoType:
		it, err := apUtil.ResolveActivityPubLink(postedActivity.Object)
		if err != nil {
			return err
		}
		follow, err := activitypub.ToActivity(it)
		if err != nil {
			return c.NoContent(http.StatusAccepted)
		}
		postedActivity.Object = follow
		if follow.Type != activitypub.FollowType || !follow.Object.GetID().Equal(bot.ID) {
			return c.NoContent(http.StatusAccepted)
		}
		if !postedActivity.Actor.GetID().Equal(follow.Actor.GetID()) {
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "Actor of Undo activity must be same to actor of object of activity")
		}
		return bc.handleUnfollow(c, bot, postedActivity)
	default:
		return c.NoContent(http.StatusAccepted)
	}
}

func (bc BotController) handleReply(c *echo.Context, bot *activitypub.Actor, call *activitypub.Activity) error {
	reply, to, err := bc.buc.Reply(c.Request().Context(), bot, call)
	if err != nil {
		if reply != nil {
			err2 := bc.buc.CancelReply(c.Request().Context(), reply)
			if err2 != nil {
				return errors.Join(err, err2)
			}
		}
		return err
	}
	if reply != nil && to != nil {
		err = bc.postActivity(c.Request().Context(), reply, to)
		if err != nil {
			return err
		}
	}
	return c.NoContent(http.StatusAccepted)
}

func (bc BotController) handleFollow(c *echo.Context, bot *activitypub.Actor, follow *activitypub.Activity) error {
	accept, to, err := bc.buc.AcceptFollowing(c.Request().Context(), bot, follow)
	if err != nil {
		return err
	}
	if accept != nil && to != nil {
		err = bc.postActivity(c.Request().Context(), accept, to)
		if err != nil {
			return err
		}
	}
	return c.NoContent(http.StatusAccepted)
}

func (bc BotController) handleUnfollow(c *echo.Context, bot *activitypub.Actor, unfollow *activitypub.Activity) error {
	_, err := bc.buc.Unfollow(c.Request().Context(), bot, unfollow)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusAccepted)
}

func (bc BotController) GetEndPoints(c *echo.Context) error {
	if strings.HasSuffix(c.Request().URL.Path, "/inbox") {
		return echo.NewHTTPError(http.StatusNotFound, "That endpoint was not found")
	}
	item, err := bc.buc.GetAny(c.Request().Context(), config.LocalOrigin().AddPath(c.Request().URL.Path))
	if apErrors.IsNotFound(err) {
		return echo.NewHTTPError(http.StatusNotFound, "That endpoint was not found")
	}
	return jldUtil.JSONLDResponse(c, http.StatusOK, item)
}

type ActivityBinder struct{}

func (ab ActivityBinder) Bind(c *echo.Context, target any) error {
	trgt, ok := target.(*activitypub.Activity)
	if !ok {
		return errors.New("ActivityBinder only supports binding to *activitypub.Activity")
	}
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	compacted, err := jldUtil.JSONCompact(body)
	if err != nil {
		return err
	}
	it, err := activitypub.UnmarshalJSON(compacted)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "failed to convert request body to activity")
	}
	activity, err := activitypub.ToActivity(it)
	if err != nil {
		return err
	}
	*trgt = *activity
	return nil
}

func (bc BotController) verifyRequest(req *http.Request, actor *activitypub.Actor) error {
	if apUtil.VerifyActivityRequest(req, actor) != nil {
		return echo.NewHTTPError(http.StatusForbidden, "Request not signed")
	}
	return nil
}

func (bc BotController) postActivity(ctx context.Context, activity *activitypub.Activity, to *activitypub.Actor) error {
	if activity.Actor == nil {
		return errors.New("actor of activity is nil")
	}
	b, err := jldUtil.JSONLDMarshal(activity)
	if err != nil {
		return err
	}
	actor, err := bc.buc.GetByUserName(ctx, schema.IDToUsername(activity.Actor.GetID()))
	if err != nil {
		return err
	}
	_, err = apUtil.PostActivityPub(actor, to.Inbox.GetLink().String(), string(b))
	return err
}
