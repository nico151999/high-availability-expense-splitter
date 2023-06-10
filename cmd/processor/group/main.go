package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nico151999/high-availability-expense-splitter/internal/processor/group"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func main() {
	log := logging.GetLogger().Named("groupProcessor")
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)

	rpProcessor, err := group.NewGroupProcessor(
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)))
	if err != nil {
		log.Panic("failed creating group processor", logging.Error(err))
	}

	if cancel, err := rpProcessor.Process(ctx); err != nil {
		log.Panic("failed processing group-related events", logging.Error(err))
	} else {
		defer func() {
			ctx, c := context.WithTimeout(ctx, 5*time.Second)
			defer c()
			if err := cancel(ctx); err != nil {
				log.Panic("failed canceling group processing", logging.Error(err))
			}
		}()
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Processing group-related events...")
	<-terminate
}
