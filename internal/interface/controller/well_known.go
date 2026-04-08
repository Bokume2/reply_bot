package controller

import (
	"encoding/xml"
	"net/http"
	"sync"

	"github.com/Bokume2/reply_bot/internal/infrastructure/config"
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
		h = webfinger.New(lw.Dev(), storage.WebFingerStorage())
	})
	return h
}

type wfHandler interface {
	HandleWebFinger(w http.ResponseWriter, r *http.Request)
}

func (wkc WellKnownController) GetHostMeta(c *echo.Context) error {
	content, err := xml.MarshalIndent(HostMetaData{
		Xmlns: "http://docs.oasis-open.org/ns/xri/xrd-1.0",
		Link: HostMetaLink{
			Rel:      "lrdd",
			Type:     "application/xrd+xml",
			Template: config.LocalOrigin().String() + webfinger.WellKnownWebFingerPath + "?resource={uri}",
		},
	}, "", "  ")
	if err != nil {
		return err
	}
	c.Response().Header().Set(echo.HeaderContentType, "application/xrd+xml")
	return c.String(http.StatusOK, xml.Header+string(content))
}

type HostMetaData struct {
	XMLName xml.Name     `xml:"XRD"`
	Xmlns   string       `xml:"xmlns,attr"`
	Link    HostMetaLink `xml:"Link"`
}

type HostMetaLink struct {
	Rel      string `xml:"rel,attr"`
	Type     string `xml:"type,attr"`
	Template string `xml:"template,attr"`
}
