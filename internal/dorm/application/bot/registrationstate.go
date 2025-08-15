package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RegistrationState struct{}

func (s *RegistrationState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	context.Telegram.DeleteMessage(chatID, context.LastMessageID)
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	err := context.UserService.RegisterUser(update.Message.Text, update.Message.From.ID)
	if err != nil {
		messageId, err := context.Telegram.SendMessage(chatID, message.RegistrationFail)
		if err != nil || messageId == 0 {
			return
		}
		context.LastMessageID = messageId
		return
	}
	messageId, err := context.Telegram.SendMessageWithReplyKeyboard(
		chatID,
		message.RegistrationSuccess,
		keyboard.MainMenu,
	)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
	context.SetState(&MainState{})
}

func (s *RegistrationState) GetName() string {
	return "WaitingForFullNameState"
}
