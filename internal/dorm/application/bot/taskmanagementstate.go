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
	chatID := s.AckCallbackAndChatID(context, update)

	switch update.CallbackQuery.Data {
	case message.ConfirmExecution:
		dutyTasks, err := s.getDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingDutyTasks, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.ConfirmExecutionState, keyboard.BuildConfirmExecutionKeyboard(dutyTasks), &ConfirmExecutionState{})
		return
	case message.AssignTask:
		areas, err := context.AreaService.GetAllAreas()
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingArea, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.AreaSelectionState, keyboard.BuildAreaKeyboard(areas), &AreaSelectionState{})
		return
	case message.UnassignTask:
		dutyTasksReadable, err := s.getDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingDutyTasks, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.TaskUnassignmentState, keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable), &TaskUnassignmentState{})
		return
	}
}

func (s *TaskManagementState) getDutyTasksReadable(context *Bot, update *tgbotapi.Update) ([]model.DutyTaskReadable, error) {
	user, err := context.UserService.GetUser(update.CallbackQuery.From.ID)
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
