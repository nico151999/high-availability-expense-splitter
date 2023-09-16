package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nico151999/high-availability-expense-splitter/internal/processor/category"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const processorName = "categoryProcessor"

func main() {
	log := logging.GetLogger().Named(processorName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)

	rpProcessor, err := category.NewCategoryProcessor(
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)),
		environment.GetDbUser(ctx),
		environment.GetDbPassword(ctx),
		fmt.Sprintf("%s:%d", environment.GetDbHost(ctx), environment.GetDbPort(ctx)),
		environment.GetDbName(ctx))
	if err != nil {
		log.Panic("failed creating category processor", logging.Error(err))
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	go func() {
		if err := rpProcessor.Process(ctx); err != nil {
			log.Panic("failed processing category-related events", logging.Error(err))
		}
	}()

	log.Info("Processing category-related events...")
	<-ctx.Done()
}
