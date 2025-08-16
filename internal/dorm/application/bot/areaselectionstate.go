package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AreaSelectionState struct {
	baseState
}

func (s *AreaSelectionState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons [][]string
	var nextState State
	switch update.CallbackQuery.Data {
	case message.Back:
		text = message.TaskManagementState
		buttons = keyboard.TaskManagementState
		nextState = &TaskManagementState{}
	}
	context.Telegram.EditMessageTextAndKeyboard(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
