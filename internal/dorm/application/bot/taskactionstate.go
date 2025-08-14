package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskActionState struct{}

func (s *TaskActionState) Handle(context *Bot, update *tgbotapi.Update) {
	context.SetState(&TaskActionState{})
}

func (s *TaskActionState) GetName() string {
	return "TaskActionState"
}
