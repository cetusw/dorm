package bot

import (
	"dorm/internal/dorm/application/service"
	"dorm/internal/dorm/infrastructure"
)

type Bot struct {
	Telegram        *infrastructure.Telegram
	UserService     *service.UserService
	AreaService     *service.AreaService
	TaskService     *service.TaskService
	DutyService     *service.DutyService
	DutyTaskService *service.DutyTaskService
	GroupService    *service.GroupService
	TeamService     *service.TeamService
	CleaningService *service.CleaningService
	State           State
	LastMessageID   int
}

func NewBot(
	telegram *infrastructure.Telegram,
	userService *service.UserService,
	areaService *service.AreaService,
	taskService *service.TaskService,
	dutyService *service.DutyService,
	dutyTaskService *service.DutyTaskService,
	groupService *service.GroupService,
	teamService *service.TeamService,
	cleaningService *service.CleaningService,
) *Bot {
	b := &Bot{
		Telegram:        telegram,
		UserService:     userService,
		AreaService:     areaService,
		TaskService:     taskService,
		DutyService:     dutyService,
		DutyTaskService: dutyTaskService,
		GroupService:    groupService,
		TeamService:     teamService,
		CleaningService: cleaningService,
	}
	b.SetState(&StartState{})
	return b
}

func (b *Bot) SetState(state State) {
	b.State = state
}
