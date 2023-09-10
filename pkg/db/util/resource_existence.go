package util

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/nico151999/high-availability-expense-splitter/pkg/logging"
	"github.com/nico151999/high-availability-expense-splitter/pkg/logging/otel"
	"github.com/rotisserie/eris"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ error = (*ResourceNotFoundError)(nil)

type ResourceNotFoundError struct {
	ResourceName string
	ResourceId   string
}

func (e ResourceNotFoundError) Error() string {
	return fmt.Sprintf("could not find %s resource with ID %s", e.ResourceName, e.ResourceId)
}

type protoWithId interface {
	proto.Message
	GetId() string
}

var ErrSelectResource = eris.New("could not select resource")

func CheckResourceExists[T protoWithId](ctx context.Context, db bun.IDB, id string) (T, error) {
	var model T
	model = reflect.New(reflect.TypeOf(model).Elem()).Interface().(T)
	modelReflect := model.ProtoReflect()
	modelDescriptor := modelReflect.Descriptor()
	modelName := string(modelDescriptor.Name())
	log := otel.NewOtelLogger(
		ctx,
		logging.FromContext(ctx).With(
			logging.String(
				"resource",
				modelName,
			),
		),
	)
	modelReflect.Set(
		modelDescriptor.Fields().ByName("id"),
		protoreflect.ValueOfString(id),
	)
	if err := db.NewSelect().Model(model).WherePK().Limit(1).Scan(ctx); err != nil {
		if eris.Is(err, sql.ErrNoRows) {
			msg := "resource not found"
			log.Debug(msg, logging.Error(err))
			return model, ResourceNotFoundError{
				ResourceName: modelName,
				ResourceId:   id,
			}
		}
		log.Error("failed getting person", logging.Error(err))
		return model, ErrSelectResource
	}
	return model, nil
}
