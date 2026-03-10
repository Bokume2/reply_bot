package controller

import (
	"errors"
	"net/http"
	domainErrors "reply_bot/internal/domain/errors"
	"reply_bot/internal/infrastructure/external"
	"reply_bot/internal/interface/schema"
	"reply_bot/internal/usecase"
	"reply_bot/internal/utils"

	"github.com/go-ap/activitypub"
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
	return utils.JSONLDResponse(c, http.StatusOK, bot)
}

func (bc BotController) GetOutBox(c *echo.Context) error {
	outbox, err := bc.buc.GetOutBox(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	return utils.JSONLDResponse(c, http.StatusOK, outbox)
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
		err = bc.postActivity(reply, to)
		if err != nil {
			return err
		}
	}
	return c.NoContent(http.StatusNoContent)
}

type ActivityBinder struct{}

func (ab ActivityBinder) Bind(c *echo.Context, item *activitypub.Item) error {
	r := c.Request()
	buf := make([]byte, r.ContentLength)
	var len int64 = 0
	for len < r.ContentLength {
		l, _ := r.Body.Read(buf[len:])
		len += int64(l)
	}
	i, err := activitypub.UnmarshalJSON(buf[:len])
	if err != nil {
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "failed to convert request body to activity")
	}
	*item = i
	return nil
}

func (bc BotController) postActivity(activity *activitypub.Activity, to *activitypub.Actor) error {
	if activity.Actor == nil {
		return errors.New("actor of activity is nil")
	}
	b, err := utils.JSONLDMarshal(activity)
	if err != nil {
		return err
	}
	username := schema.IDToUsername(activity.Actor.GetID())
	_, err = external.PostActivityPub(utils.PKeyPath(username), to.Inbox.GetLink().String(), string(b))
	return err
}
