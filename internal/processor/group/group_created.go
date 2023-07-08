package group

import (
	"context"

	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *groupProcessor) groupCreated(ctx context.Context, req *groupv1.GroupCreated) error {
	log := logging.FromContext(ctx)
	log.Info("processing group.GroupCreated event",
		logging.String("name", req.GetName()),
		logging.String("groupId", req.GetGroupId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message like sending a project created email and publish an event telling what was done (e.g. project creation email sent)
	return nil
}
