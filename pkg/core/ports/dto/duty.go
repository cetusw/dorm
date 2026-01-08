package dto

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"fmt"
	"time"
)

type DutyViewModel struct {
	TeamName      string
	TeamOrder     int
	SpreadsheetID string
	DutyName      string
	Start         time.Time
	End           time.Time
	UsersStats    []UserStats
	Tasks         []TaskViewModel
}

func NewDutyViewModel(
	d *duty.Duty,
	team *structure.Team,
	group *structure.Group,
	tasks []TaskViewModel,
	usersStats []UserStats,
) DutyViewModel {
	return DutyViewModel{
		TeamName:      team.Name(),
		TeamOrder:     team.Order(),
		SpreadsheetID: group.SpreadsheetID(),
		DutyName:      fmt.Sprintf("%s-%s", d.Start().Format("02.01"), d.End().Format("02.01")),
		Start:         d.Start(),
		End:           d.End(),
		UsersStats:    usersStats,
		Tasks:         tasks,
	}
}
