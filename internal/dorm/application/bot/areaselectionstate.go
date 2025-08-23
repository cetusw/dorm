package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
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
		lastDuty, err := context.DutyService.GetLastDuty()
		if err != nil {
			log.Printf("ERROR: failed to get last duty: %v", err)
			return
		}
		tasks, err := context.DutyTaskService.GetUnassignedDutyTasksReadableByAreaIDAndDutyID(areaID, lastDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to get unassigned tasks for area %d: %v", areaID, err)
			return
		}

		s.EditInlineAndGo(
			context,
			chatID,
			message.TaskAssignmentState,
			keyboard.BuildTaskAssignmentKeyboard(tasks),
			&TaskAssignmentState{AreaID: areaID},
		)
	}
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
