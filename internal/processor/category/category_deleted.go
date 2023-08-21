package category

import (
	"context"

	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *categoryProcessor) categoryDeleted(ctx context.Context, req *categoryv1.CategoryDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing category.CategoryDeleted event",
		logging.String("categoryId", req.GetCategoryId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
