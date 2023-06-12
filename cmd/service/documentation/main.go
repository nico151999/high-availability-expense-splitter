package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func main() {
	log := logging.GetLogger().Named("documentationSvc")
	ctx := logging.IntoContext(context.Background(), log)

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	log.Info("Running Documentation service...")
	<-ctx.Done()
}
