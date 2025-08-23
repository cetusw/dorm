package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type TaskUnassignmentState struct {
	baseState
}

func (s *TaskUnassignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
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

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixUnassign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixUnassign)
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
		err = context.DutyTaskService.SetDutyTaskAssigneeID(nil, taskUUID, currentDuty.DutyID)
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
			message.TaskUnassignmentState,
			keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable),
			&TaskUnassignmentState{},
		)
		err = context.CleaningService.UpdateCurrentSheet(currentDuty)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (s *TaskUnassignmentState) GetName() string {
	return "TaskUnassignmentState"
}
