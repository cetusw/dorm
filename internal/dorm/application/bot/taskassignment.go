package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type TaskAssignmentState struct {
	baseState
	AreaID int
}

func (s *TaskAssignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)
	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back {
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
		return nil
	}

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixAssign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixAssign)
		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		err = s.AssignTask(context, chatID, taskUUID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		unassignedTasks, err := s.GetUnassignedTasks(context, chatID, s.AreaID)
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
			fmt.Sprintf(message.TaskAssignmentState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
			keyboard.BuildTaskAssignmentKeyboard(unassignedTasks),
			&TaskAssignmentState{AreaID: s.AreaID},
		)
		user, err := context.UserService.GetUserByTelegramID(chatID)
		if err != nil {
			return err
		}
		err = context.CleaningService.UpdateCurrentSheet(*user)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
	}

	return nil
}

func (s *TaskAssignmentState) GetName() string {
	return "TaskAssignmentState"
}
