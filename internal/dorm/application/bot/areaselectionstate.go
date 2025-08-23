package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
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

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixArea) {
		areaID, err := s.GetAreaID(callbackData)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		duty, err := s.GetCurrentDuty(context)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		tasks, err := s.GetUnassignedTasks(context, areaID, duty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		user, err := s.GetUser(context, chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		userPoints, pointsPerUser, err := s.GetPointsSummary(context, *user.TeamID, user.UserID, duty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return nil
		}

		s.EditInlineAndGo(
			context,
			chatID,
			fmt.Sprintf(message.TaskAssignmentState, userPoints, pointsPerUser),
			keyboard.BuildTaskAssignmentKeyboard(tasks),
			&TaskAssignmentState{AreaID: areaID},
		)
	}

	return nil
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
