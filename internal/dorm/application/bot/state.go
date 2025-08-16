package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type State interface {
	Handle(context *Bot, update *tgbotapi.Update)
	HandleCallback(context *Bot, update *tgbotapi.Update)
	GetName() string
}
