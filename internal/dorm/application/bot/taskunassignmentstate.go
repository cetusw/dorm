package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskUnassignmentState struct {
	baseState
}

func (s *TaskUnassignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons tgbotapi.InlineKeyboardMarkup
	var nextState State
	switch update.CallbackQuery.Data {
	case message.Back:
		text = message.TaskManagementState
		buttons = keyboard.BuildTaskManagementKeyboard()
		nextState = &TaskManagementState{}
	}
	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *TaskUnassignmentState) GetName() string {
	return "TaskUnassignmentState"
}
