package bot

import (
	"dorm/internal/common/keyboard"
	"dorm/internal/common/message"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type ConfirmExecutionState struct {
	baseState
}

func (s *ConfirmExecutionState) HandleCallback(context *Bot, update *tgbotapi.Update) {
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

	if strings.HasPrefix(callbackData, keyboard.CallbackPrefixConfirm) {
		taskID := strings.TrimPrefix(callbackData, keyboard.CallbackPrefixConfirm)
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
		err = context.DutyTaskService.CompleteDutyTaskByDutyIDAndTaskID(currentDuty.DutyID, taskUUID)
		if err != nil {
			log.Printf("ERROR: failed to complete task: %v", err)
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
			message.ConfirmExecutionState,
			keyboard.BuildConfirmExecutionKeyboard(dutyTasksReadable),
			&ConfirmExecutionState{},
		)
		err = context.CleaningService.UpdateCurrentSheet(currentDuty)
		if err != nil {
			log.Printf("ERROR: failed to update sheet: %v", err)
			return
		}
	}
}

func (s *ConfirmExecutionState) GetName() string {
	return "ConfirmExecutionState"
}
