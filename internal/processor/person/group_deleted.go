package person

import (
	"context"
	"database/sql"

	personv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/person/v1"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	personprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/person/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *personProcessor) groupDeleted(ctx context.Context, req *groupprocv1.GroupDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("groupId", req.GetId()))
	log.Info("processing group.GroupDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var people []*personv1.Person
		if err := tx.NewDelete().Model(&people).Where("group_id = ?", req.GetId()).Returning("id").Scan(ctx); err != nil {
			log.Error("failed deleting people related to deleted group", logging.Error(err))
			return errDeletePeople
		}

		g, _ := errgroup.WithContext(ctx)
		for _, c := range people {
			person := c
			g.Go(func() error {
				marshalled, err := proto.Marshal(&personprocv1.PersonDeleted{
					Id: person.Id,
				})
				if err != nil {
					log.Error("failed marshalling person deleted event", logging.Error(err))
					return errMarshalPersonDeleted
				}
				if err := rpProcessor.natsClient.Publish(environment.GetPersonDeletedSubject(req.GetId(), person.Id), marshalled); err != nil {
					log.Error("failed publishing person deleted event", logging.Error(err))
					return errPublishPersonDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
