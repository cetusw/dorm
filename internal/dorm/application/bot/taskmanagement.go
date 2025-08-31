package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskManagementState struct {
	baseState
}

func (s *TaskManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)

	switch update.CallbackQuery.Data {
	case message.ConfirmExecution:
		dutyTasks, err := s.GetUncompletedDutyTasksView(context, update.CallbackQuery.From.ID)
		if len(dutyTasks) == 0 {
			s.EditInlineAndGo(context, chatID, message.NoTasksToConfirm, keyboard.BuildTaskManagementKeyboard(), &TaskManagementState{})
			return nil
		}
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		progress, err := s.ComputeUserProgress(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}

		s.EditInlineAndGo(
			context,
			chatID,
			fmt.Sprintf(message.ConfirmExecutionState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
			keyboard.BuildConfirmExecutionKeyboard(dutyTasks),
			&ConfirmExecutionState{},
		)
	case message.AssignTask:
		unassignedAreas, err := s.GetUnassignedAreas(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		progress, err := s.ComputeUserProgress(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.EditInlineAndGo(
			context,
			chatID,
			fmt.Sprintf(message.AreaSelectionState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
			keyboard.BuildAreaKeyboard(unassignedAreas),
			&AreaSelectionState{},
		)
	case message.UnassignTask:
		uncompletedTasks, err := s.GetUncompletedDutyTasksView(context, update.CallbackQuery.From.ID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		progress, err := s.ComputeUserProgress(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		s.EditInlineAndGo(
			context,
			chatID,
			fmt.Sprintf(message.TaskUnassignmentState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
			keyboard.BuildTaskUnassignmentKeyboard(uncompletedTasks),
			&TaskUnassignmentState{},
		)
	}

	return nil
}

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
