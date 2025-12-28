package bot

import (
	"dorm/pkg/common/keyboard"
	"dorm/pkg/common/message"
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AreaSelectionState struct {
	baseState
}

func (s *AreaSelectionState) HandleCallback(context *Bot, update *tgbotapi.Update) error {
	chatID := s.AckCallbackAndChatID(context, update)
	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back { // TODO: подумать над тем, чтобы передавать данные в класс для совершение перехода назад
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

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixArea) {
		areaID, err := s.GetCallbackID(callbackData, keyboard.CallbackPrefixArea)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		tasks, err := s.GetUnassignedTasks(context, chatID, areaID)
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
			keyboard.BuildTaskAssignmentKeyboard(tasks),
			&TaskAssignmentState{AreaID: areaID},
		)
	}

	return nil
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
