package dto

import (
	"fmt"

	"github.com/google/uuid"

	"dorm/pkg/core/domain/catalog"
	"dorm/pkg/core/domain/duty"
	"dorm/pkg/core/domain/user"
)

type TaskViewModel struct {
	ID               uuid.UUID
	Title            string
	AreaName         string
	AreaID           int
	AreaFloor        int
	Cost             int
	IsDone           bool
	IsAssignedToUser bool
	Executor         string
}

func NewTaskViewModel(
	dt *duty.DutyTask,
	def *catalog.TaskDefinition,
	area *catalog.Area,
	assignee *user.User,
	currentUserID uuid.UUID,
) TaskViewModel {
	executorName := "Никто"
	if assignee != nil {
		executorName = fmt.Sprintf("%s %s.", assignee.FirstName(), string([]rune(assignee.LastName())[0]))
	}

	return TaskViewModel{
		ID:               dt.ID(),
		Title:            def.Title(),
		AreaID:           area.ID(),
		AreaName:         area.Name(),
		AreaFloor:        area.Floor(),
		Cost:             def.Cost(),
		IsDone:           dt.IsCompleted(),
		IsAssignedToUser: dt.AssigneeID() != nil && *dt.AssigneeID() == currentUserID,
		Executor:         executorName,
	}
}
