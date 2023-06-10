package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func main() {
	log := logging.GetLogger().Named("groupSvc")
	_ = logging.IntoContext(context.Background(), log)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Running Group service...")
	<-terminate
}
