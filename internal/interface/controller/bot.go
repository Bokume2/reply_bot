package controller

import (
	"errors"
	"net/http"
	domainErrors "reply_bot/internal/domain/errors"
	"reply_bot/internal/usecase"

	"github.com/go-ap/activitypub"
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
	if c.Request().Header.Get(echo.HeaderContentType) != "application/activity+json" &&
		c.Request().Header.Get(echo.HeaderContentType) != "application/ld+json;profile=\"https://www.w3.org/ns/activitystreams\"" {
		c.Response().Header().Del(echo.HeaderContentType)
		return echo.NewHTTPError(http.StatusUnsupportedMediaType, "expected application/activity+json")
	}
	var ab ActivityBinder
	item := new(activitypub.Item)
	if err := ab.Bind(c, item); err != nil {
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	_, err := bc.buc.Reply(c.Request().Context(), c.Param("username"), item)
	if err != nil {
		c.Response().Header().Del(echo.HeaderContentType)
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

type ActivityBinder struct{}

func (ab ActivityBinder) Bind(c *echo.Context, i any) error {
	r := c.Request()
	buf := make([]byte, r.ContentLength)
	r.Body.Read(buf)
	item, err := activitypub.UnmarshalJSON(buf)
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
