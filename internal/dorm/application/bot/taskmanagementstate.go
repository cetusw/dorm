package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"dorm/internal/dorm/application/model"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TaskManagementState struct {
	baseState
}

func (s *TaskManagementState) HandleCallback(context *Bot, update *tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	callbackQueryID := update.CallbackQuery.ID
	context.Telegram.AnswerCallbackQuery(callbackQueryID, "")
	var text string
	var buttons tgbotapi.InlineKeyboardMarkup
	var nextState State

	switch update.CallbackQuery.Data {
	case message.ConfirmExecution:
		dutyTasks, err := s.getDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			text = message.ErrorWhileGettingDutyTasks
			nextState = &MainState{}
			break
		}
		text = message.ConfirmExecutionState
		buttons = keyboard.BuildConfirmExecutionKeyboard(dutyTasks)
		nextState = &ConfirmExecutionState{}
	case message.AssignTask:
		areas, err := context.AreaService.GetAllAreas()
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			text = message.ErrorWhileGettingArea
			nextState = &MainState{}
			break
		}
		text = message.AreaSelectionState
		buttons = keyboard.BuildAreaKeyboard(areas)
		nextState = &AreaSelectionState{}
	case message.UnassignTask:
		dutyTasksReadable, err := s.getDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			text = message.ErrorWhileGettingDutyTasks
			nextState = &MainState{}
			break
		}
		text = message.TaskUnassignmentState
		buttons = keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable)
		nextState = &TaskUnassignmentState{}
	}

	context.Telegram.EditMessageWithMarkup(chatID, context.LastMessageID, text, buttons)
	context.SetState(nextState)
}

func (s *TaskManagementState) getDutyTasksReadable(context *Bot, update *tgbotapi.Update) ([]model.DutyTaskReadable, error) {
	user, err := context.UserService.GetUser(update.Message.From.ID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get user while getting duty tasks: %v", err)
	}
	lastDuty, err := context.DutyService.GetLastDuty()
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get last duty tasks: %v", err)
	}
	dutyTasksReadable, err := context.DutyTaskService.GetDutyTasksReadableByAssigneeIDAndDutyID(user.UserID, lastDuty.DutyID)
	if err != nil {
		return nil, fmt.Errorf("ERROR: failed to get areas for keyboard: %v", err)
	}

	return dutyTasksReadable, nil
}

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
