package router

import (
	"github.com/Bokume2/reply_bot/internal/interface/controller"

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
	botsRouter.Use(controller.BotsMiddleware)
	botsRouter.GET("/:username", r.botController.GetByUserName)
	// specified bot
	botRouter := botsRouter.Group("/:username")
	botRouter.GET("/outbox", r.botController.GetOutBox)
	botRouter.POST("/inbox", r.botController.PostInBox)
	botRouter.GET("/*", r.botController.GetEndPoints)

	// nodeinfo
	r.echo.GET(webfinger.NodeInfoDiscoverPath, r.nodeIndoController.GetNodeInfoDiscover)
	r.echo.GET(controller.NodeInfoPath, r.nodeIndoController.GetNodeInfoContent)

	// webfinger
	r.echo.GET(webfinger.WellKnownWebFingerPath, r.wellKnownController.GetWebfinger)

	return r.echo
}
