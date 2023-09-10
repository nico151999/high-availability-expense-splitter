package expensestake

import (
	"context"

	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expensestakeProcessor) expensestakeUpdated(ctx context.Context, req *expensestakev1.ExpenseStakeUpdated) error {
	log := logging.FromContext(ctx)
	log.Info("processing expensestake.ExpenseStakeUpdated event",
		logging.String("expensestakeId", req.GetId()))
	// TODO: actually process message like sending a project updated notification and publish an event telling what was done (e.g. project updated notification sent)
	return nil
}
