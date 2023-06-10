package model

import (
	groupv1 "github.com/nico151999/high-availability-expense-splitter/gen/lib/go/common/group/v1"
)

type Group struct {
	groupv1.GroupProperties
	GroupId string
}
