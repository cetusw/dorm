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
		{message.Back},
	}
	TeamManagementMenu = [][]string{
		{message.TeamMembers},
		{message.DutySchedule},
		{message.Back},
	}
	ReservationManagementMenu = [][]string{
		{message.ReserveTask},
		{message.ConfirmExecution},
		{message.Back},
	}
	BackMenu = [][]string{
		{message.Back},
	}
)
