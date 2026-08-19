package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"github.com/example/federated-analytics-coordinator/internal/platform/config"
	"github.com/example/federated-analytics-coordinator/internal/platform/httpapi"
	"github.com/example/federated-analytics-coordinator/internal/platform/store"
	service "github.com/example/federated-analytics-coordinator/internal/protocol/application"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, e := config.Load()
	if e != nil {
		slog.Error("configuration error", "error", e)
		os.Exit(1)
	}
	ids := func() string { b := make([]byte, 16); _, _ = rand.Read(b); return hex.EncodeToString(b) }
	api := httpapi.New(cfg, service.New(store.New(), ids))
	server := &http.Server{Addr: cfg.Address, Handler: api.Handler(), ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		slog.Info("coordinator started", "address", cfg.Address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("listen failed", "error", err)
			os.Exit(1)
		}
	}()
	signalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	<-signalCtx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if e := server.Shutdown(ctx); e != nil {
		slog.Error("shutdown failure", "error", e)
		os.Exit(1)
	}
}
