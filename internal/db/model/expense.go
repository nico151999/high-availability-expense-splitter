package model

import (
	expensev1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/expense/v1"
)

type Expense struct {
	expensev1.Expense
	Timestamp *Timestamp
}

func NewExpense(expense *expensev1.Expense) *Expense {
	var name *string
	if expense != nil {
		name = expense.Name
	}
	return &Expense{
		Expense: expensev1.Expense{
			Id:         expense.GetId(),
			GroupId:    expense.GetGroupId(),
			Name:       name,
			ById:       expense.GetById(),
			CurrencyId: expense.GetCurrencyId(),
		},
		Timestamp: NewTimestamp(expense.GetTimestamp()),
	}
}

func (e *Expense) IntoProtoExpense() *expensev1.Expense {
	e.Expense.Timestamp = e.Timestamp.IntoProtoTimestamp()
	return &e.Expense
}
