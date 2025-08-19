package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	SetupLogger()
	SetupConfig()
	SetupMqtt()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,    // ^C
		syscall.SIGTERM, // kill default
	)
	defer stop()

	slog.Info("waiting for exit signal")

	<-ctx.Done()
	slog.Info("shutting down")
}
