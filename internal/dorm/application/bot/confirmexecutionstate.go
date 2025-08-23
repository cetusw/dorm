package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ConfirmExecutionState struct {
	baseState
}

func (s *ConfirmExecutionState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := s.AckCallbackAndChatID(context, update)
	switch update.CallbackQuery.Data {
	case message.Back:
		s.EditInlineAndGo(
			context,
			chatID,
			message.TaskManagementState,
			keyboard.BuildTaskManagementKeyboard(),
			&TaskManagementState{},
		)
	}
}

func (s *ConfirmExecutionState) GetName() string {
	return "ConfirmExecutionState"
}
