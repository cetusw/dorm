package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"fmt"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type TaskAssignmentState struct {
	baseState
	AreaID int
}

func (s *TaskAssignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := s.AckCallbackAndChatID(context, update)
	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back {
		areas, err := context.AreaService.GetAllAreas()
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingArea, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.AreaSelectionState, keyboard.BuildAreaKeyboard(areas), &AreaSelectionState{})
		return
	}

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixAssign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixAssign)
		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			log.Printf("ERROR: failed to parse task ID from callback data '%s': %v", callbackData, err)
			return
		}
		currentDuty, err := context.DutyService.GetCurrentDuty()
		if err != nil {
			log.Printf("ERROR: failed to get last duty: %v", err)
			return
		}
		user, err := context.UserService.GetUser(chatID)
		if err != nil {
			log.Printf("ERROR: failed to get user while getting duty tasks: %v", err)
			return
		}
		err = context.DutyTaskService.SetDutyTaskAssigneeIDByDutyID(user.UserID, taskUUID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to set current duty task assignee: %v", err)
			return
		}
		tasks, err := context.DutyTaskService.GetUnassignedDutyTasksReadableByAreaIDAndDutyID(s.AreaID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to get unassigned tasks for area %d: %v", s.AreaID, err)
			return
		}
		userPoints, pointsPerUser, err := s.GetPointsSummary(context, *user.TeamID, user.UserID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to compute points summary: %v", err)
			return
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
			log.Printf("ERROR: failed to update sheet: %v", err)
			return
		}
	}
}

func (s *TaskAssignmentState) GetName() string {
	return "TaskAssignmentState"
}
