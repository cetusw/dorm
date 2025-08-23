package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type TaskUnassignmentState struct {
	baseState
}

func (s *TaskUnassignmentState) HandleCallback(context *Bot, update *tgbotapi.Update) {
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

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixUnassign) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixUnassign)
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
		err = context.DutyTaskService.SetDutyTaskAssigneeIDByDutyID(nil, taskUUID, currentDuty.DutyID)
		if err != nil {
			log.Printf("ERROR: failed to set current duty task assignee: %v", err)
			return
		}
		dutyTasksReadable, err := s.GetUncompletedDutyTasksReadable(context, update)
		if err != nil {
			log.Println(err)
			s.SendReplyAndGo(context, chatID, message.ErrorWhileGettingDutyTasks, keyboard.BuildMainStateKeyboard(), &MainState{})
			return
		}

		s.EditInlineAndGo(
			context,
			chatID,
			message.TaskUnassignmentState,
			keyboard.BuildTaskUnassignmentKeyboard(dutyTasksReadable),
			&TaskUnassignmentState{},
		)
		err = context.CleaningService.UpdateCurrentSheet(currentDuty)
		if err != nil {
			log.Printf("ERROR: failed to update sheet: %v", err)
			return
		}
	}
}

func (s *TaskUnassignmentState) GetName() string {
	return "TaskUnassignmentState"
}
