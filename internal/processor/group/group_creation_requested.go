package group

import (
	"context"

	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
)

func (rpProcessor *groupProcessor) groupCreationRequested(ctx context.Context, req *groupv1.GroupCreationRequested) error {
	log := logging.FromContext(ctx)
	log.Info("processing group.GroupCreationRequested event",
		logging.String("name", req.GetName()),
		logging.String("groupId", req.GetGroupId()),
		logging.String("requestorEmail", req.GetRequestorEmail()))
	// TODO: actually process message
	// TODO: publish event telling that greoup was created
	return nil
}
