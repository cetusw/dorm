package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ProfileManagementState struct{}

func (s *ProfileManagementState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""
	buttons := keyboard.MainMenu
	context.Telegram.DeleteMessage(chatID, context.LastMessageID)
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	switch update.Message.Text {
	default:
		text = message.Back
		buttons = keyboard.MainMenu
		context.SetState(&MainState{})
	}
	messageId, err := context.Telegram.SendMessageWithReplyKeyboard(chatID, text, buttons)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
}

func (s *ProfileManagementState) GetName() string {
	return "HelpState"
}
