package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"fmt"
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
		uncompletedTasks, err := s.GetUncompletedDutyTasksView(context, update.CallbackQuery.From.ID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}

		s.EditInlineAndGo(
			context,
			chatID,
			message.TaskManagementState,
			keyboard.BuildTaskManagementKeyboard(uncompletedTasks),
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
		err = s.ConfirmTask(context, taskUUID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		dutyTasks, err := s.GetUncompletedDutyTasksView(context, update.CallbackQuery.From.ID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		progress, err := s.ComputeUserProgress(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		uncompletedTasks, err := s.GetUncompletedDutyTasksView(context, update.CallbackQuery.From.ID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}

		if len(dutyTasks) == 0 {
			s.EditInlineAndGo(context, chatID, message.AllTasksConfirmed, keyboard.BuildTaskManagementKeyboard(uncompletedTasks), &TaskManagementState{})
		} else {
			s.EditInlineAndGo(
				context,
				chatID,
				fmt.Sprintf(message.ConfirmExecutionState, progress.UserPoints, progress.UserConfirmedPoints, progress.UserRequiredPoints),
				keyboard.BuildConfirmExecutionKeyboard(dutyTasks),
				&ConfirmExecutionState{},
			)
		}
		user, err := context.UserService.GetUser(chatID)
		if err != nil {
			return err
		}
		err = context.CleaningService.UpdateCurrentSheet(*user)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ConfirmExecutionState) GetName() string {
	return "ConfirmExecutionState"
}
