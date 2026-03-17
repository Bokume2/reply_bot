package controller

import (
	"context"
	"errors"
	"io"
	"net/http"
	domainErrors "reply_bot/internal/domain/errors"
	"reply_bot/internal/infrastructure/config"
	externalAP "reply_bot/internal/infrastructure/external/activitypub"
	externalJSONLD "reply_bot/internal/infrastructure/external/jsonld"
	"reply_bot/internal/interface/schema"
	"reply_bot/internal/usecase"
	"strings"

	"github.com/go-ap/activitypub"
	apErrors "github.com/go-ap/errors"
	"github.com/go-ap/jsonld"
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
	return externalJSONLD.JSONLDResponse(c, http.StatusOK, bot)
}

func (bc BotController) GetOutBox(c *echo.Context) error {
	outbox, err := bc.buc.GetOutBox(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	return externalJSONLD.JSONLDResponse(c, http.StatusOK, outbox)
}

func (bc BotController) PostInBox(c *echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) != "application/activity+json" && c.Request().Header.Get(echo.HeaderContentType) != jsonld.ContentType {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "expected application/activity+json")
	}
	var ab ActivityBinder
	item := new(activitypub.Item)
	if err := ab.Bind(c, item); err != nil {
		return err
	}
	reply, to, err := bc.buc.Reply(c.Request().Context(), c.Param("username"), item)
	if err != nil {
		if reply != nil {
			item := activitypub.Item(reply)
			bc.buc.CancelReply(c.Request().Context(), &item)
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

func (bc BotController) GetEndPoints(c *echo.Context) error {
	if strings.HasSuffix(c.Request().URL.Path, "/inbox") {
		return echo.NewHTTPError(http.StatusNotFound, "That endpoint was not found")
	}
	item, err := bc.buc.GetAny(c.Request().Context(), config.LOCAL_ORIGIN.AddPath(c.Request().URL.Path))
	if apErrors.IsNotFound(err) {
		return echo.NewHTTPError(http.StatusNotFound, "That endpoint was not found")
	}
	return externalJSONLD.JSONLDResponse(c, http.StatusOK, item)
}

type ActivityBinder struct{}

func (ab ActivityBinder) Bind(c *echo.Context, item *activitypub.Item) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	i, err := activitypub.UnmarshalJSON(body)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "failed to convert request body to activity")
	}
	*item = i
	return nil
}

func (bc BotController) postActivity(ctx context.Context, activity *activitypub.Activity, to *activitypub.Actor) error {
	if activity.Actor == nil {
		return errors.New("actor of activity is nil")
	}
	b, err := externalJSONLD.JSONLDMarshal(activity)
	if err != nil {
		return err
	}
	actor, err := bc.buc.GetByUserName(ctx, schema.IDToUsername(activity.Actor.GetID()))
	if err != nil {
		return err
	}
	_, err = externalAP.PostActivityPub(actor, to.Inbox.GetLink().String(), string(b))
	return err
}
