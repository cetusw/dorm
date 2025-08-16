package keyboard

import "dorm/internal/common/message"

var (
	MainState = [][]string{
		{message.Tasks},
		{message.Team},
		{message.Payment},
		{message.Profile},
	}
	TaskManagementState = [][]string{
		{message.ConfirmExecution},
		{message.AssignTask},
		{message.UnassignTask},
	}
	ConfirmExecutionState = [][]string{
		{message.Back},
		{"// TODO: список назначенных задач"},
	}
	AreaSelectionState = [][]string{
		{message.Back},
		{"// TODO: список зон"},
	}
	TaskAssignmentState = [][]string{
		{message.Back},
		{"// TODO: список незакреплённых задач"},
	}
	TaskUnassignmentState = [][]string{
		{message.Back},
		{"// TODO: список назначенных задач"},
	}
	TeamManagementState = [][]string{
		{message.TeamMembers},
		{message.DutySchedule},
	}
	BackMenu = [][]string{
		{message.Back},
	}
)
