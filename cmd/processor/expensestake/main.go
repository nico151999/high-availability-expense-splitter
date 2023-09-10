package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nico151999/high-availability-expense-splitter/internal/processor/expensestake"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

const processorName = "expensestakeProcessor"

func main() {
	log := logging.GetLogger().Named(processorName)
	ctx := logging.IntoContext(context.Background(), log)

	// ensure mandatory environment variables are set
	environment.GetNatsServerHost(ctx)
	environment.GetNatsServerPort(ctx)

	rpProcessor, err := expensestake.NewExpenseStakeProcessor(
		fmt.Sprintf("%s:%d",
			environment.GetNatsServerHost(ctx),
			environment.GetNatsServerPort(ctx)),
		environment.GetDbUser(ctx),
		environment.GetDbPassword(ctx),
		fmt.Sprintf("%s:%d", environment.GetDbHost(ctx), environment.GetDbPort(ctx)),
		environment.GetDbName(ctx))
	if err != nil {
		log.Panic("failed creating expensestake processor", logging.Error(err))
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// TODO: generally switch to context-based processing
	if cancel, err := rpProcessor.Process(ctx); err != nil {
		log.Panic("failed processing expensestake-related events", logging.Error(err))
	} else {
		defer func() {
			ctx, c := context.WithTimeout(ctx, 5*time.Second)
			defer c()
			if err := cancel(ctx); err != nil {
				log.Panic("failed canceling expensestake processing", logging.Error(err))
			}
		}()
	}

	log.Info("Processing expensestake-related events...")
	<-ctx.Done()
}
