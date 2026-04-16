package dto

import (
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/structure"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type DutyViewModel struct {
	TeamID        uuid.UUID
	TeamName      string
	TeamOrder     int
	TeamColor     string
	SpreadsheetID string
	DutyName      string
	Start         time.Time
	End           time.Time
	UsersStats    []*UserStats
	Tasks         []TaskViewModel
}

func NewDutyViewModel(
	duty *duty.Duty,
	team *structure.Team,
	group *structure.Group,
	usersStats []*UserStats,
	tasks []TaskViewModel,
) DutyViewModel {
	return DutyViewModel{
		TeamID:        team.ID(),
		TeamName:      team.Name(),
		TeamOrder:     team.Order(),
		TeamColor:     team.Color(),
		SpreadsheetID: group.SpreadsheetID(),
		DutyName:      formatDutyName(duty),
		Start:         duty.Start(),
		End:           duty.End(),
		UsersStats:    usersStats,
		Tasks:         tasks,
	}
}

func formatDutyName(duty *duty.Duty) string {
	return fmt.Sprintf(
		"%s-%s",
		duty.Start().Format("02.01"),
		duty.End().Format("02.01"),
	)
}
