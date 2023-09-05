package category

import (
	"context"
	"database/sql"

	categoryv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/category/v1"
	categoryprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/category/v1"
	groupprocv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/processor/group/v1"
	"github.com/nico151999/high-availability-expense-splitter/pkg/environment"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/uptrace/bun"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/proto"
)

func (rpProcessor *categoryProcessor) groupDeleted(ctx context.Context, req *groupprocv1.GroupDeleted) error {
	log := logging.FromContext(ctx).With(logging.String("groupId", req.GetId()))
	log.Info("processing group.GroupDeleted event")

	return rpProcessor.dbClient.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		var categories []*categoryv1.Category
		if err := tx.NewDelete().Model(&categories).Where("group_id = ?", req.GetId()).Returning("id").Scan(ctx); err != nil {
			log.Error("failed deleting categories related to deleted group", logging.Error(err))
			return errDeleteCategories
		}

		g, _ := errgroup.WithContext(ctx)
		for _, c := range categories {
			category := c
			g.Go(func() error {
				marshalled, err := proto.Marshal(&categoryprocv1.CategoryDeleted{
					Id: category.Id,
				})
				if err != nil {
					log.Error("failed marshalling category deleted event", logging.Error(err))
					return errMarshalCategoryDeleted
				}
				if err := rpProcessor.natsClient.Publish(environment.GetCategoryDeletedSubject(req.GetId(), category.Id), marshalled); err != nil {
					log.Error("failed publishing category deleted event", logging.Error(err))
					return errPublishCategoryDeleted
				}
				return nil
			})
		}
		return g.Wait()
	})
}
