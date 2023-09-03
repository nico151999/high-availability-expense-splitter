package person

import (
	"context"

	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *personProcessor) personCreated(ctx context.Context, req *personv1.PersonCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing person.PersonCreated event",
		logging.String("name", req.GetName()),
		logging.String("personId", req.GetId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created notification and publish an event telling what was done (e.g. project creation notification sent)
	return nil
}
