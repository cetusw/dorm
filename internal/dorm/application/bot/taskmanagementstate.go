package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskManagementState struct {
	baseState
}

func (s *TaskManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)

	switch update.CallbackQuery.Data {
	case message.ConfirmExecution:
		dutyTasks, err := s.GetUncompletedDutyTasksView(context, update)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.EditInlineAndGo(context, chatID, message.ConfirmExecutionState, keyboard.BuildConfirmExecutionKeyboard(dutyTasks), &ConfirmExecutionState{})
	case message.AssignTask:
		currentDuty, err := context.DutyService.GetCurrentDuty()
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		areas, err := context.AreaService.GetUnassignedAreasByDutyID(currentDuty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.EditInlineAndGo(context, chatID, message.AreaSelectionState, keyboard.BuildAreaKeyboard(areas), &AreaSelectionState{})
	case message.UnassignTask:
		dutyTasksReadable, err := s.GetUncompletedDutyTasksView(context, update)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.EditInlineAndGo(context, chatID, message.TaskUnassignmentState, keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable), &TaskUnassignmentState{})
	}

	return nil
}

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
