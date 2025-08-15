package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MainState struct{}

func (s *MainState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""
	buttons := keyboard.MainMenu
	context.Telegram.DeleteMessage(chatID, context.LastMessageID)
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	switch update.Message.Text {
	case message.Tasks:
		text = message.Tasks
		buttons = keyboard.TaskManagementMenu
		context.SetState(&TaskManagementState{})
	case message.Team:
		text = message.Team
		buttons = keyboard.TeamManagementMenu
		context.SetState(&TeamManagementState{})
		break
	case message.Payment:
		text = message.PaymentLink
		buttons = keyboard.BackMenu
		context.SetState(&PaymentManagementState{})
	case message.Profile:
		text = message.Profile // TODO: сделать профиль
		buttons = keyboard.BackMenu
		context.SetState(&ProfileManagementState{})
	}
	messageId, err := context.Telegram.SendMessageWithReplyKeyboard(chatID, text, buttons)
	if err != nil || messageId == 0 {
		return
	}
	context.LastMessageID = messageId
}

func (s *MainState) GetName() string {
	return "MainState"
}
