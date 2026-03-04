package controller

import (
	"net/http"
	"reply_bot/internal/domain/errors"
	"reply_bot/internal/usecase"

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
		if err == errors.ErrBotNotFound {
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
	return c.JSON(http.StatusOK, bot)
}

func (bc BotController) GetOutBox(c *echo.Context) error {
	outBox, err := bc.buc.GetOutBox(c.Request().Context(), c.Param("username"))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, outBox)
}
