package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartState struct{}

func (s *StartState) Handle(context *Bot, update *tgbotapi.Update) {
	user, err := context.UserService.GetUser(update.Message.From.ID)
	if err != nil {
		log.Println(err)
		context.Telegram.SendMessage(update.Message.Chat.ID, message.RegistrationFail)
	}
	if user != nil {
		context.Telegram.SendMessageWithReplyKeyboard(update.Message.Chat.ID, message.Welcome, keyboard.TaskAction)
		context.SetState(&TaskActionState{})
		return
	}
	context.Telegram.SendMessage(update.Message.Chat.ID, message.Registration)
	context.SetState(&WaitingForFullNameState{})
}

func (s *StartState) GetName() string {
	return "StartState"
}
