// Команда server запускает REST API с регистрацией и авторизацией.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/saydullobro/tutorial.go/rest-advanced/internal/auth"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/config"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/database"
	"github.com/saydullobro/tutorial.go/rest-advanced/internal/router"
)

func main() {
	if err := run(); err != nil {
		slog.Error("завершение с ошибкой", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(db, auth.NewTokenManager(cfg.JWTSecret, cfg.TokenTTL)),
		ReadHeaderTimeout: 5 * time.Second, // защита от медленных клиентов
	}

	// ctx отменяется по Ctrl+C или SIGTERM (его шлёт Docker/Kubernetes).
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		slog.Info("сервер запущен", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	// Graceful shutdown: даём текущим запросам до 10 секунд на завершение.
	slog.Info("останавливаемся")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	if err := <-errCh; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
