package person

import (
	"context"

	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *personProcessor) personUpdated(ctx context.Context, req *personv1.PersonUpdated) error {
	log := logging.FromContext(ctx)
	log.Info("processing person.PersonUpdated event",
		logging.String("personId", req.GetId()))
	// TODO: actually process message like sending a project updated notification and publish an event telling what was done (e.g. project updated notification sent)
	return nil
}
