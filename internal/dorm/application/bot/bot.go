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
	DutyService     *service.DutyService
	DutyTaskService *service.DutyTaskService
	CleaningService *service.CleaningService
	State           State
}

func NewBot(
	telegram *infrastructure.Telegram,
	userService *service.UserService,
	areaService *service.AreaService,
	taskService *service.TaskService,
	dutyService *service.DutyService,
	dutyTaskService *service.DutyTaskService,
	cleaningService *service.CleaningService,
) *Bot {
	b := &Bot{
		Telegram:        telegram,
		UserService:     userService,
		AreaService:     areaService,
		TaskService:     taskService,
		DutyService:     dutyService,
		DutyTaskService: dutyTaskService,
		CleaningService: cleaningService,
	}
	b.SetState(&StartState{})
	return b
}

func (b *Bot) SetState(state State) {
	b.State = state
}
