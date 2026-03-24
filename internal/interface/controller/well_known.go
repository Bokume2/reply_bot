package controller

import (
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage"

	"git.sr.ht/~mariusor/lw"
	"github.com/go-ap/webfinger"
	"github.com/labstack/echo/v5"
)

type WellKnownController struct{}

func NewWellKnownController() *WellKnownController {
	return &WellKnownController{}
}

func (wkc WellKnownController) GetWebfinger(c *echo.Context) error {
	webfinger.New(lw.Prod(), storage.WebFingerStorage).HandleWebFinger(c.Response(), c.Request())
	return nil
}
