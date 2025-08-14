package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WaitingForFullNameState struct{}

func (s *WaitingForFullNameState) Handle(context *Bot, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID

	err := context.UserService.RegisterUser(update.Message.Text, update.Message.From.ID)
	if err != nil {
		context.Telegram.SendMessage(chatId, message.RegistrationFail)
		return
	}

	context.Telegram.SendMessageWithReplyKeyboard(chatId, message.RegistrationSuccess, keyboard.TaskAction)

	context.SetState(&TaskActionState{})
}

func (s *WaitingForFullNameState) GetName() string {
	return "WaitingForFullNameState"
}
