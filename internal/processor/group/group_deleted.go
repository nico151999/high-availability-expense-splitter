package group

import (
	"context"

	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *groupProcessor) groupDeleted(ctx context.Context, req *groupv1.GroupDeleted) error {
	log := logging.FromContext(ctx)
	log.Info("processing group.GroupDeleted event",
		logging.String("groupId", req.GetId()))
	// TODO: actually process message like sending a project deleted notification and publish an event telling what was done (e.g. project deleted notification sent)
	return nil
}
