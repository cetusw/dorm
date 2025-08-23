package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AreaSelectionState struct {
	baseState
}

func (s *AreaSelectionState) HandleCallback(context *Bot, update *tgbotapi.Update) {
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
		return
	}

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixArea) {
		areaIDStr := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixArea)
		areaID, err := strconv.Atoi(areaIDStr)
		if err != nil {
			log.Printf("ERROR: invalid area ID in callback: %v", err)
			return
		}
		currentDuty, err := context.DutyService.GetLastDuty()
		if err != nil {
			log.Printf("ERROR: failed to get last duty: %v", err)
			return
		}
		tasks, err := context.DutyTaskService.GetUnassignedDutyTasksReadableByAreaIDAndDutyID(areaID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to get unassigned tasks for area %d: %v", areaID, err)
			return
		}
		user, err := context.UserService.GetUser(chatID)
		if err != nil {
			log.Printf("ERROR: failed to get user while getting duty tasks: %v", err)
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
			&TaskAssignmentState{AreaID: areaID},
		)
	}
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
