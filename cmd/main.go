package main

import (
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/infrastructure/router"
	"reply_bot/internal/infrastructure/template"
	"reply_bot/internal/interface/controller"

	"github.com/labstack/echo/v5"
)

func main() {
	config.LoadEnv()
	template.LoadTemplate()

	nic := &controller.NodeInfoController{}
	wkc := &controller.WellKnownController{}

	e := router.NewRouter(echo.New(), nic, wkc).Setup()

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
