package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskManagementState struct {
	baseState
}

func (s *TaskManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons [][]string
	var nextState State

	switch update.CallbackQuery.Data {
	case message.ConfirmExecution:
		text = message.ConfirmExecutionState
		buttons = keyboard.ConfirmExecutionState // TODO: сформировать список задач
		nextState = &ConfirmExecutionState{}
	case message.AssignTask:
		text = message.AreaSelectionState
		buttons = keyboard.AreaSelectionState
		nextState = &AreaSelectionState{}
	case message.UnassignTask:
		text = message.TaskUnassignmentState
		buttons = keyboard.TaskUnassignmentState
		nextState = &TaskUnassignmentState{}
	}

	context.Telegram.EditMessageTextAndKeyboard(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
