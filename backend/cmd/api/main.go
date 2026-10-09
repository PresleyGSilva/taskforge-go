package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/PresleyGSilva/taskforge-go/internal/config"
	"github.com/PresleyGSilva/taskforge-go/internal/db"
	"github.com/PresleyGSilva/taskforge-go/internal/httpapi"
	"github.com/PresleyGSilva/taskforge-go/internal/store"
	"github.com/PresleyGSilva/taskforge-go/internal/ws"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := db.Migrate(ctx, pool, "migrations"); err != nil {
		if err2 := db.Migrate(ctx, pool, "backend/migrations"); err2 != nil {
			slog.Error("migrate failed", "err", err)
			os.Exit(1)
		}
	}

	srv := &httpapi.Server{Store: store.NewPgx(pool), JWTSecret: cfg.JWTSecret, Hub: ws.NewHub()}
	httpSrv := &http.Server{Addr: ":" + cfg.Port, Handler: httpapi.NewRouter(srv)}

	go func() {
		slog.Info("listening", "port", cfg.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("serve failed", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	sctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(sctx)
	slog.Info("shutdown complete")
}
