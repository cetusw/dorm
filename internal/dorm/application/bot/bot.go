package bot

import (
	"dorm/internal/dorm/application/service"
	"dorm/internal/dorm/infrastructure"
)

type Bot struct {
	Telegram    *infrastructure.Telegram
	UserService *service.UserService
	State       State
}

func NewBot(telegram *infrastructure.Telegram, userService *service.UserService) *Bot {
	b := &Bot{
		Telegram:    telegram,
		UserService: userService,
	}
	b.SetState(&StartState{})
	return b
}

func (b *Bot) SetState(state State) {
	b.State = state
}
