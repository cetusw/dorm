package bot

import (
	"dorm/pkg/common/keyboard"
	"dorm/pkg/common/message"
	"fmt"
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
			fmt.Sprintf(message.TaskManagementState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
			keyboard.BuildTaskManagementKeyboard(uncompletedTasks),
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
		err = s.UnassignTask(context, chatID, taskUUID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
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
		user, err := context.UserService.GetUserByTelegramID(chatID)
		if err != nil {
			return err
		}
		err = context.CleaningService.UpdateCurrentSheet(*user)
		if err != nil {
			return nil
		}
	}

	return nil
}

func (s *TaskUnassignmentState) GetName() string {
	return "TaskUnassignmentState"
}
