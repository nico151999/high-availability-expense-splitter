package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nico151999/high-availability-expense-splitter/internal/processor/group"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const processorName = "groupProcessor"

func main() {
	log := logging.GetLogger().Named(processorName)
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

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	go func() {
		if err := rpProcessor.Process(ctx); err != nil {
			log.Panic("failed processing group-related events", logging.Error(err))
		}
	}()

	log.Info("Processing group-related events...")
	<-ctx.Done()
}
