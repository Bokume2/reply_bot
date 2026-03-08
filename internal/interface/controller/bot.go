package controller

import (
	"errors"
	"fmt"
	"net/http"
	domainErrors "reply_bot/internal/domain/errors"
	"reply_bot/internal/usecase"
	"reply_bot/internal/utils"
	"strings"

	"github.com/go-ap/activitypub"
	"github.com/go-ap/httpsig"
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
		c.Response().Header().Set(echo.HeaderContentType, "application/activity+json")
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
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	return c.JSON(http.StatusOK, bot)
}

func (bc BotController) GetOutBox(c *echo.Context) error {
	outBox, err := bc.buc.GetOutBox(c.Request().Context(), c.Param("username"))
	if err != nil {
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	return c.JSON(http.StatusOK, outBox)
}

func (bc BotController) PostInBox(c *echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) != "application/activity+json" && c.Request().Header.Get(echo.HeaderContentType) != jsonld.ContentType {
		c.Response().Header().Del(echo.HeaderContentType)
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "expected application/activity+json")
	}
	var ab ActivityBinder
	item := new(activitypub.Item)
	if err := ab.Bind(c, item); err != nil {
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	reply, to, err := bc.buc.Reply(c.Request().Context(), c.Param("username"), item)
	if err != nil {
		if reply != nil {
			item := activitypub.Item(reply)
			bc.buc.CancelReply(c.Request().Context(), &item)
		}
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	if reply != nil {
		err = bc.postActivity(reply, to)
		if err != nil {
			c.Response().Header().Del(echo.HeaderContentType)
			return err
		}
	}
	return c.NoContent(http.StatusNoContent)
}

type ActivityBinder struct{}

func (ab ActivityBinder) Bind(c *echo.Context, i any) error {
	r := c.Request()
	buf := make([]byte, r.ContentLength)
	var len int64 = 0
	for len < r.ContentLength {
		l, _ := r.Body.Read(buf[len:])
		len += int64(l)
	}
	item, err := activitypub.UnmarshalJSON(buf[:len])
	if err != nil {
		return err
	}
	switch k := i.(type) {
	case *activitypub.Item:
		*k = item
	default:
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "failed to convert request body to activity")
	}
	return nil
}

func (bc BotController) postActivity(activity *activitypub.Activity, to *activitypub.Actor) error {
	b, err := activitypub.MarshalJSON(activity)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", to.Inbox.GetLink().String(), strings.NewReader(string(b)))
	if err != nil {
		return err
	}
	req.Header.Set(echo.HeaderContentType, "application/activity+json")
	bot, err := activitypub.ToActor(activity.Actor)
	if err != nil {
		return err
	}
	key, err := utils.ReadPrivKey(fmt.Sprintf("storage/cred/%s.key", bot.PreferredUsername))
	if err != nil {
		return err
	}
	signer := httpsig.NewRSASHA256Signer("signer", key, nil)
	err = signer.Sign(req)
	if err != nil {
		return err
	}
	_, err = new(http.Client).Do(req)
	return err
}
