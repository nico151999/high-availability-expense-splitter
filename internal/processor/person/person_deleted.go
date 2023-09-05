package person

import (
	"context"

	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *personProcessor) personDeleted(ctx context.Context, req *personv1.PersonDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing person.PersonDeleted event",
		logging.String("personId", req.GetId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
