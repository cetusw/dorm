package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskManagementState struct{}

func (s *TaskManagementState) Handle(context *Bot, update *tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := ""
	buttons := keyboard.MainMenu
	context.Telegram.DeleteMessage(chatID, context.LastMessageID)
	context.Telegram.DeleteMessage(chatID, update.Message.MessageID)
	switch update.Message.Text {
	case message.Reservation:
		text = message.Reservation
		buttons = keyboard.ReservationManagementMenu
		context.SetState(&ReservationManagementState{})
	case message.ConfirmExecution:
		text = message.ConfirmExecution
		buttons = keyboard.BackMenu // TODO: сформировать список задач и вставить его в клавиатуру
		context.SetState(&ConfirmationManagementState{})
	case message.Back:
		text = message.Back
		buttons = keyboard.MainMenu
		context.SetState(&MainState{})
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

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
