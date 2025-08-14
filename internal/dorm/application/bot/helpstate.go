package bot

import (
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpState struct{}

func (s *HelpState) Handle(context *Bot, update *tgbotapi.Update) {
	context.Telegram.SendMessage(update.Message.Chat.ID, message.Help)
}

func (s *HelpState) GetName() string {
	return "HelpState"
}
