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
	outbox, err := bc.buc.GetOutBox(c.Request().Context(), c.Param("username"))
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
	reply, to, err := bc.buc.Reply(c.Request().Context(), c.Param("username"), postedActivity)
	if err != nil {
		if reply != nil {
			bc.buc.CancelReply(c.Request().Context(), reply)
		}
		return err
	}
	if reply != nil && to != nil {
		err = bc.verifyRequest(c.Request(), to)
		if err != nil {
			return err
		}
		err = bc.postActivity(c.Request().Context(), reply, to)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusAccepted)
	}
	accept, to, err := bc.buc.AcceptFollowing(c.Request().Context(), c.Param("username"), postedActivity)
	if err != nil {
		return err
	}
	if accept != nil && to != nil {
		err = bc.verifyRequest(c.Request(), to)
		if err != nil {
			return err
		}
		err = bc.postActivity(c.Request().Context(), accept, to)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusAccepted)
	}
	done, err := bc.buc.Unfollow(c.Request().Context(), c.Param("username"), postedActivity)
	if err != nil {
		return err
	}
	if done {
		return c.NoContent(http.StatusAccepted)
	}
	return echo.NewHTTPError(http.StatusUnprocessableEntity, "That activity is not accepted by this server")
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
