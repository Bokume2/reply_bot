package main

import (
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/infrastructure/router"
	"reply_bot/internal/infrastructure/storage"
	"reply_bot/internal/infrastructure/storage/repository/bot"
	"reply_bot/internal/infrastructure/template"
	"reply_bot/internal/interface/controller"
	"reply_bot/internal/usecase"

	"github.com/labstack/echo/v5"
)

func main() {
	config.LoadEnv()
	template.LoadTemplate()
	storage.InitStorage()

	nic := &controller.NodeInfoController{}
	wkc := &controller.WellKnownController{}
	buc := usecase.NewBotUseCase(bot.NewBotRepository(storage.DataStore))
	bc := controller.NewBotController(buc)

	e := router.NewRouter(echo.New(), bc, nic, wkc).Setup()

	if err := e.Start(":3000"); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
