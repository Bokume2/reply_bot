package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bokume2/reply_bot/internal/infrastructure/router"
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage"
	"github.com/Bokume2/reply_bot/internal/infrastructure/storage/repository/bot"
	"github.com/Bokume2/reply_bot/internal/interface/controller"
	"github.com/Bokume2/reply_bot/internal/usecase"

	"github.com/labstack/echo/v5"
)

func main() {
	br := bot.NewBotRepository(storage.DataStore)

	buc := usecase.NewBotUseCase(br)

	nic := controller.NewNodeInfoController()
	wkc := controller.NewWellKnownController()
	bc := controller.NewBotController(buc)

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
