package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
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
		dutyTasks, err := s.GetUncompletedDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingDutyTasks, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.ConfirmExecutionState, keyboard.BuildConfirmExecutionKeyboard(dutyTasks), &ConfirmExecutionState{})
		return
	case message.AssignTask:
		currentDuty, err := context.DutyService.GetCurrentDuty()
		if err != nil {
			log.Printf("ERROR: failed to get last duty: %v", err)
			return
		}
		areas, err := context.AreaService.GetUnassignedAreasByDutyID(currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to get areas for keyboard: %v", err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingArea, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.AreaSelectionState, keyboard.BuildAreaKeyboard(areas), &AreaSelectionState{})
		return
	case message.UnassignTask:
		dutyTasksReadable, err := s.GetUncompletedDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingDutyTasks, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}
		s.EditInlineAndGo(context, chatID, message.TaskUnassignmentState, keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable), &TaskUnassignmentState{})
		return
	}
}

func (s *TaskManagementState) GetName() string {
	return "TaskManagementState"
}
