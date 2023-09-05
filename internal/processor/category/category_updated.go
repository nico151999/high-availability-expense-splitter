package category

import (
	"context"

	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *categoryProcessor) categoryUpdated(ctx context.Context, req *categoryv1.CategoryUpdated) error {
	log := logging.FromContext(ctx)
	log.Info("processing category.CategoryUpdated event",
		logging.String("categoryId", req.GetId()))
	// TODO: actually process message like sending a project updated notification and publish an event telling what was done (e.g. project updated notification sent)
	return nil
}
