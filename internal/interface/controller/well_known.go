package controller

import (
	"net/http"
	"sync"

	"github.com/Bokume2/reply_bot/internal/infrastructure/storage"

	"git.sr.ht/~mariusor/lw"
	"github.com/go-ap/webfinger"
	"github.com/labstack/echo/v5"
)

var (
	h       wfHandler
	wkcOnce sync.Once
)

type WellKnownController struct{}

func NewWellKnownController() *WellKnownController {
	return &WellKnownController{}
}

func (wkc WellKnownController) GetWebfinger(c *echo.Context) error {
	handler().HandleWebFinger(c.Response(), c.Request())
	return nil
}

func handler() wfHandler {
	wkcOnce.Do(func() {
		h = webfinger.New(lw.Dev(), storage.WebFingerStorage)
	})
	return h
}

type wfHandler interface {
	HandleWebFinger(w http.ResponseWriter, r *http.Request)
}
