package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartState struct{}

func (s *StartState) Handle(context *Bot, update *tgbotapi.Update) {
	chatId := update.Message.Chat.ID
	context.Telegram.DeleteMessage(chatId, update.Message.MessageID)
	user, err := context.UserService.GetUser(update.Message.From.ID)
	if err != nil {
		messageId, err := context.Telegram.SendMessage(chatId, message.RegistrationFail)
		if err != nil || messageId == 0 {
			return
		}
		context.LastMessageID = messageId
	}
	if user == nil {
		messageId, err := context.Telegram.SendMessage(chatId, message.Registration)
		if err != nil || messageId == 0 {
			return
		}
		context.LastMessageID = messageId
		context.SetState(&RegistrationState{})
		return
	}
	messageId, err := context.Telegram.SendMessageWithReplyKeyboard(
		chatId,
		message.Welcome,
		keyboard.MainMenu,
	)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
	context.SetState(&MainState{})
}

func (s *StartState) HandleCallback(context *Bot, update *tgbotapi.Update) {}

func (s *StartState) GetName() string {
	return "StartState"
}
