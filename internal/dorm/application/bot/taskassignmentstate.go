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
		return nil
	}

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixAssign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixAssign)
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
		user, err := context.UserService.GetUser(chatID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		err = context.DutyTaskService.SetDutyTaskAssigneeID(&user.UserID, taskUUID, currentDuty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		tasks, err := context.DutyTaskService.GetUnassignedDutyTasksView(s.AreaID, currentDuty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}
		userPoints, pointsPerUser, err := s.GetPointsSummary(context, *user.TeamID, user.UserID, currentDuty.DutyID)
		if err != nil {
			s.SendReplyAndGo(context, chatID, message.Error, keyboard.BuildMainStateKeyboard(), &MainState{})
			return err
		}

		s.EditInlineAndGo(
			context,
			chatID,
			fmt.Sprintf(message.TaskAssignmentState, userPoints, pointsPerUser),
			keyboard.BuildTaskAssignmentKeyboard(tasks),
			&TaskAssignmentState{AreaID: s.AreaID},
		)
		err = context.CleaningService.UpdateCurrentSheet(currentDuty)
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
