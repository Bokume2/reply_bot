package controller

import (
	"net/http"

	"github.com/Bokume2/reply_bot/internal/infrastructure/template"

	"github.com/labstack/echo/v5"
)

type NodeInfoController struct{}

func NewNodeInfoController() *NodeInfoController {
	return &NodeInfoController{}
}

func (nic NodeInfoController) GetNodeInfoContent(c *echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.String(http.StatusOK, template.NodeInfo)
}
