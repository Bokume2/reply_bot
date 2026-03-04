package router

import (
	"reply_bot/internal/interface/controller"

	"github.com/go-ap/webfinger"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Router struct {
	echo                *echo.Echo
	botController       *controller.BotController
	nodeIndoController  *controller.NodeInfoController
	wellKnownController *controller.WellKnownController
}

func NewRouter(e *echo.Echo, bc *controller.BotController, nic *controller.NodeInfoController, wkc *controller.WellKnownController) *Router {
	return &Router{
		echo:                e,
		botController:       bc,
		nodeIndoController:  nic,
		wellKnownController: wkc,
	}
}

func (r *Router) Setup() *echo.Echo {
	r.echo.Use(middleware.RequestLogger())
	r.echo.Use(middleware.Recover())

	// public files
	r.echo.Static("/public", "public")

	// bots
	botsRouter := r.echo.Group("/bots")
	botsRouter.Use(controller.CheckBotExistance)
	botsRouter.GET("/:username", r.botController.GetByUserName)

	// nodeinfo
	r.echo.GET("/.well-known/nodeinfo", r.wellKnownController.GetNodeInfo)
	r.echo.GET("/nodeinfo/2.1", r.nodeIndoController.GetNodeInfoContent)

	// webfinger
	r.echo.GET(webfinger.WellKnownWebFingerPath, r.wellKnownController.GetWebfinger)

	return r.echo
}
