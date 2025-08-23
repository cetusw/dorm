package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type RegistrationState struct {
	baseState
}

func (s *RegistrationState) Handle(context *Bot, update *tgbotapi.Update) error {
	chatID := update.Message.Chat.ID
	context.Telegram.ClearDialogue(chatID, update.Message.MessageID, context.LastMessageID)
	err := context.UserService.RegisterUser(update.Message.Text, update.Message.From.ID)
	if err != nil {
		messageID, err := context.Telegram.SendMessage(chatID, message.RegistrationFail)
		if err != nil || messageID == 0 {
			return err
		}
		context.LastMessageID = messageID
		return err
	}
	s.SendReplyAndGo(
		context,
		chatID,
		message.RegistrationSuccess,
		keyboard.BuildMainStateKeyboard(),
		&MainState{},
	)

	return nil
}

func (s *RegistrationState) GetName() string {
	return "RegistrationState"
}
