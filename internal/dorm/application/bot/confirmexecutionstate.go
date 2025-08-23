package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type ConfirmExecutionState struct {
	baseState
}

func (s *ConfirmExecutionState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)
	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back {
		s.EditInlineAndGo(
			context,
			chatID,
			message.TaskManagementState,
			keyboard.BuildTaskManagementKeyboard(),
			&TaskManagementState{},
		)
		return nil
	}

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixConfirm) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixConfirm)
		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		currentDuty, err := context.DutyService.GetCurrentDuty()
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		err = context.DutyTaskService.CompleteDutyTaskByDutyIDAndTaskID(currentDuty.DutyID, taskUUID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		dutyTasksReadable, err := s.GetUncompletedDutyTasksView(context, update)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}

		s.EditInlineAndGo(
			context,
			chatID,
			message.ConfirmExecutionState,
			keyboard.BuildConfirmExecutionKeyboard(dutyTasksReadable),
			&ConfirmExecutionState{},
		)
		err = context.CleaningService.UpdateCurrentSheet(currentDuty)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ConfirmExecutionState) GetName() string {
	return "ConfirmExecutionState"
}
