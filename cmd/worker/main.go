package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("worker stopped")
			return
		case <-ticker.C:
			slog.Info("checking protocol deadlines")
		}
	}
}
