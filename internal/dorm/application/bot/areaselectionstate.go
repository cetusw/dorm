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
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")

	var text string
	var buttons tgbotapi.InlineKeyboardMarkup
	var nextState State

	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back {
		text = message.TaskManagementState
		buttons = keyboard.BuildTaskManagementKeyboard()
		nextState = &TaskManagementState{}
		return
	} else if strings.HasPrefix(callbackData, keyboard.CallbackPrefixArea) {
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

		text = message.TaskAssignmentState
		buttons = keyboard.BuildTaskAssignmentKeyboard(tasks)

		nextState = &TaskAssignmentState{AreaID: areaID}
	}

	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *AreaSelectionState) GetName() string {
	return "AreaSelectionState"
}
