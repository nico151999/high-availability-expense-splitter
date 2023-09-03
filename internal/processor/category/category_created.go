package category

import (
	"context"

	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *categoryProcessor) categoryCreated(ctx context.Context, req *categoryv1.CategoryCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing category.CategoryCreated event",
		logging.String("name", req.GetName()),
		logging.String("categoryId", req.GetId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
