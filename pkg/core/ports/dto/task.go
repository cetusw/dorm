package dto

import (
	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
)

type TaskViewModel struct {
	ID          uuid.UUID
	Title       string
	AreaName    string
	AreaID      int
	AreaFloor   int
	Cost        int
	Assignee    *UserStats
	IsCompleted bool
	IsVerified  bool
}

func NewTaskViewModel(
	dt *duty.DutyTask,
	def *catalog.TaskDefinition,
	area *catalog.Area,
	assignee *UserStats,
) TaskViewModel {
	return TaskViewModel{
		ID:          dt.ID(),
		Title:       def.Title(),
		AreaID:      area.ID(),
		AreaName:    area.Name(),
		AreaFloor:   area.Floor(),
		Cost:        def.Cost(),
		Assignee:    assignee,
		IsCompleted: dt.CompletionDate() != nil,
		IsVerified:  dt.VerificationDate() != nil,
	}
}
