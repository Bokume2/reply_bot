package controller

import (
	"net/http"
	"reply_bot/internal/infrastructure/template"

	"github.com/labstack/echo/v5"
)

type WellKnownController struct{}

func (wkc WellKnownController) GetNodeInfo(c *echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(http.StatusOK, template.WellKnownNodeInfo)
}
