package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nico151999/high-availability-expense-splitter/internal/processor/expensecategoryrelation"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const processorName = "expensecategoryrelationProcessor"

func main() {
	log := logging.GetLogger().Named(processorName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)

	rpProcessor, err := expensecategoryrelation.NewExpenseCategoryRelationProcessor(
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)),
		environment.GetDbUser(ctx),
		environment.GetDbPassword(ctx),
		fmt.Sprintf("%s:%d", environment.GetDbHost(ctx), environment.GetDbPort(ctx)),
		environment.GetDbName(ctx))
	if err != nil {
		log.Panic("failed creating expensecategoryrelation processor", logging.Error(err))
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	go func() {
		if err := rpProcessor.Process(ctx); err != nil {
			log.Panic("failed processing expensecategoryrelation-related events", logging.Error(err))
		}
	}()

	log.Info("Processing expensecategoryrelation-related events...")
	<-ctx.Done()
}
