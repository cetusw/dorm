package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type StartState struct {
	baseState
}

func (s *StartState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	user, err := context.UserService.GetUser(update.Message.From.ID)
	if err != nil {
		messageID, err := context.Telegram.SendMessage(chatID, message.RegistrationFail)
		if err != nil || messageID == 0 {
			return
		}
		context.LastMessageID = messageID
	}
	if user == nil {
		messageID, err := context.Telegram.SendMessage(chatID, message.Registration)
		if err != nil || messageID == 0 {
			return
		}
		context.LastMessageID = messageID
		context.SetState(&RegistrationState{})
		return
	}
	messageID, err := context.Telegram.SendMessageWithMarkup(
		chatID,
		message.MainState,
		keyboard.BuildMainStateKeyboard(),
	)
	if err != nil || messageID == 0 {
		return
	}
	context.LastMessageID = messageID
	context.SetState(&MainState{})
}

func (s *StartState) GetName() string {
	return "StartState"
}
