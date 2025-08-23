package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
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
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons tgbotapi.InlineKeyboardMarkup
	var nextState State

	callbackData := update.CallbackQuery.Data

	if callbackData == message.Back {
		areas, err := context.AreaService.GetAllAreas()
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			text = message.ErrorWhileGettingArea
			nextState = &MainState{}
			return
		}
		text = message.AreaSelectionState
		buttons = keyboard.BuildAreaKeyboard(areas)
		nextState = &AreaSelectionState{}
		return
	} else if strings.HasPrefix(callbackData, keyboard.CallbackPrefixAssign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixAssign)
		taskUUID, err := uuid.Parse(taskID)
		if err != nil {
			log.Printf("ERROR: failed to parse task ID from callback data '%s': %v", callbackData, err)
			return
		}
		currentDuty, err := context.DutyService.GetLastDuty()
		if err != nil {
			log.Printf("ERROR: failed to get last duty: %v", err)
			return
		}
		err = context.CleaningService.HandleTaskAssignment(chatID, taskUUID, currentDuty)
		if err != nil {
			log.Printf("ERROR: failed to handle task assignment: %v", err)
			return
		}
		tasks, err := context.DutyTaskService.GetUnassignedDutyTasksReadableByAreaIDAndDutyID(s.AreaID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to get unassigned tasks for area %d: %v", s.AreaID, err)
			return
		}

		text = message.TaskAssignmentState
		buttons = keyboard.BuildTaskAssignmentKeyboard(tasks)

		nextState = &TaskAssignmentState{AreaID: s.AreaID}
	}

	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *TaskAssignmentState) GetName() string {
	return "TaskAssignmentState"
}
