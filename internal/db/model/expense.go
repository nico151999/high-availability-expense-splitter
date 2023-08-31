package model

import (
	"time"

	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
	"github.com/uptrace/bun"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ExpenseModel struct {
	bun.BaseModel `bun:"table:expenses"`
	*expensev1.Expense

	Timestamp time.Time
}

func NewExpense(em *expensev1.Expense) *ExpenseModel {
	return &ExpenseModel{
		Expense:   em,
		Timestamp: em.GetTimestamp().AsTime(),
	}
}

func (em *ExpenseModel) IntoExpense() *expensev1.Expense {
	e := em.Expense
	e.Timestamp = timestamppb.New(em.Timestamp)
	return e
}
