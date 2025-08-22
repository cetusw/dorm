package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskAssignmentState struct {
	baseState
	AreaID int
}

func (s *TaskAssignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons tgbotapi.InlineKeyboardMarkup
	var nextState State
	switch update.CallbackQuery.Data {
	case message.Back:
		areas, err := context.AreaService.GetAllAreas()
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			text = message.ErrorWhileGettingArea
			break
		}
		text = message.AreaSelectionState
		buttons = keyboard.BuildAreaKeyboard(areas)
		nextState = &AreaSelectionState{}
	}
	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *TaskAssignmentState) GetName() string {
	return "TaskAssignmentState"
}
