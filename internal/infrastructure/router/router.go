package router

import (
	"reply_bot/internal/interface/controller"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Router struct {
	echo                *echo.Echo
	nodeIndoController  *controller.NodeInfoController
	wellKnownController *controller.WellKnownController
}

func NewRouter(e *echo.Echo, nic *controller.NodeInfoController, wkc *controller.WellKnownController) *Router {
	return &Router{
		echo:                e,
		nodeIndoController:  nic,
		wellKnownController: wkc,
	}
}

func (r *Router) Setup() *echo.Echo {
	r.echo.Use(middleware.RequestLogger())
	r.echo.Use(middleware.Recover())

	// nodeinfo
	r.echo.GET("/.well-known/nodeinfo", r.wellKnownController.GetNodeInfo)
	r.echo.GET("/nodeinfo/2.1", r.nodeIndoController.GetNodeInfoContent)

	return r.echo
}
