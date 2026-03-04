package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"reply_bot/internal/infrastructure/config"
	"reply_bot/internal/infrastructure/router"
	"reply_bot/internal/infrastructure/storage"
	"reply_bot/internal/infrastructure/storage/repository/bot"
	"reply_bot/internal/infrastructure/template"
	"reply_bot/internal/interface/controller"
	"reply_bot/internal/usecase"
	"syscall"
	"time"

	"github.com/labstack/echo/v5"
)

func main() {
	config.LoadEnv()
	template.LoadTemplate()
	storage.InitStorage()

	nic := controller.NewNodeInfoController()
	wkc := controller.NewWellKnownController()
	bc := controller.NewBotController(usecase.NewBotUseCase(bot.NewBotRepository(storage.DataStore)))

	e := router.NewRouter(echo.New(), bc, nic, wkc).Setup()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	s := http.Server{Addr: ":3000", Handler: e}

	go func() {
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Error("failed to start server", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		e.Logger.Error("failed to stop server", "error", err)
	}
}
