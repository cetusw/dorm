package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type State interface {
	Handle(context *Bot, update *tgbotapi.Update) error
	HandleCallback(context *Bot, update *tgbotapi.Update) error
	GetName() string
}
