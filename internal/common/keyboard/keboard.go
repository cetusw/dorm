package keyboard

import "dorm/internal/common/message"

var (
	MainMenu = [][]string{
		{message.Tasks},
		{message.Team},
		{message.Payment},
		{message.Profile},
	}
	TaskManagementMenu = [][]string{
		{message.Reservation},
		{message.ConfirmExecution},
	}
	TeamManagementMenu = [][]string{
		{message.TeamMembers},
		{message.DutySchedule},
	}
	ReservationManagementMenu = [][]string{
		{message.ReserveTask},
		{message.ConfirmExecution},
	}
	BackMenu = [][]string{
		{message.Back},
	}
)
