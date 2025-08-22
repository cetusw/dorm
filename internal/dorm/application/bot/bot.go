package bot

import (
	"dorm/internal/dorm/application/service"
	"dorm/internal/dorm/infrastructure"
)

type Bot struct {
	Telegram        *infrastructure.Telegram
	LastMessageID   int
	UserService     *service.UserService
	AreaService     *service.AreaService
	TaskService     *service.TaskService
	DutyTaskService *service.DutyTaskService
	State           State
}

func NewBot(
	telegram *infrastructure.Telegram,
	userService *service.UserService,
	areaService *service.AreaService,
	taskService *service.TaskService,
	dutyTaskService *service.DutyTaskService,
) *Bot {
	b := &Bot{
		Telegram:        telegram,
		UserService:     userService,
		AreaService:     areaService,
		TaskService:     taskService,
		DutyTaskService: dutyTaskService,
	}
	b.SetState(&StartState{})
	return b
}

func (b *Bot) SetState(state State) {
	b.State = state
}
