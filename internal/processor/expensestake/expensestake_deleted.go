package expensestake

import (
	"context"

	expensestakev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/expensestake/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *expensestakeProcessor) expensestakeDeleted(ctx context.Context, req *expensestakev1.ExpenseStakeDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing expensestake.ExpenseStakeDeleted event",
		logging.String("expensestakeId", req.GetId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
