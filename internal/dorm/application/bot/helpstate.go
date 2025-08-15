package bot

import (
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type HelpState struct{}

func (s *HelpState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	messageId, err := context.Telegram.SendMessage(chatID, message.Help)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
}

func (s *HelpState) GetName() string {
	return "HelpState"
}
