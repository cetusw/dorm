package keyboard

import (
	"dorm/pkg/common/message"
	"dorm/pkg/dorm/application/model"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackPrefixArea     = "area_select_"
	CallbackPrefixAssign   = "assign_task_"
	CallbackPrefixUnassign = "unassign_task_"
	CallbackPrefixConfirm  = "confirm_task_"
)

func BuildMainStateKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return buildReplyKeyboard(
		[]buttonInfo{
			{message.Tasks, message.Tasks},
			{message.Payment, message.Payment},
			{message.Profile, message.Profile},
		},
	)
}

func BuildTaskManagementKeyboard(uncompletedTasks []model.DutyTaskView) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	if len(uncompletedTasks) == 0 {
		button := tgbotapi.NewInlineKeyboardButtonData(message.AssignTask, message.AssignTask)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))

		return tgbotapi.NewInlineKeyboardMarkup(rows...)
	}

	button := tgbotapi.NewInlineKeyboardButtonData(message.ConfirmExecution, message.ConfirmExecution)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	button = tgbotapi.NewInlineKeyboardButtonData(message.AssignTask, message.AssignTask)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	button = tgbotapi.NewInlineKeyboardButtonData(message.UnassignTask, message.UnassignTask)
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))

	return tgbotapi.NewInlineKeyboardMarkup(rows...)

}

func BuildTeamManagementKeyboard() tgbotapi.InlineKeyboardMarkup {
	return buildSimpleKeyboard(
		[]buttonInfo{
			{message.TeamMembers, message.TeamMembers},
			{message.DutySchedule, message.DutySchedule},
		},
	)
}

func BuildBackKeyboard() tgbotapi.InlineKeyboardMarkup {
	return buildSimpleKeyboard([]buttonInfo{{message.Back, message.Back}})
}

func BuildAreaKeyboard(areas []model.Area) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, area := range areas {
		callbackData := fmt.Sprintf("%s%d", CallbackPrefixArea, area.ID)
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%d этаж. %s", area.Floor, area.Name), callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildTaskAssignmentKeyboard(unassignedTasks []model.DutyTaskView) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, task := range unassignedTasks {
		callbackData := fmt.Sprintf("%s%s", CallbackPrefixAssign, task.TaskID)
		button := tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s: %d", task.TaskTitle, task.TaskCost), callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildConfirmExecutionKeyboard(assignedTasks []model.DutyTaskView) tgbotapi.InlineKeyboardMarkup {
	return buildAssignedTaskKeyboard(assignedTasks, CallbackPrefixConfirm)
}

func BuildTaskUnassignmentKeyboard(assignedTasks []model.DutyTaskView) tgbotapi.InlineKeyboardMarkup {
	return buildAssignedTaskKeyboard(assignedTasks, CallbackPrefixUnassign)
}

type buttonInfo struct {
	Text string
	Data string
}

func buildSimpleKeyboard(buttons []buttonInfo) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, btn := range buttons {
		button := tgbotapi.NewInlineKeyboardButtonData(btn.Text, btn.Data)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func buildReplyKeyboard(buttons []buttonInfo) tgbotapi.ReplyKeyboardMarkup {
	var rows [][]tgbotapi.KeyboardButton
	for _, btn := range buttons {
		button := tgbotapi.NewKeyboardButton(btn.Text)
		rows = append(rows, tgbotapi.NewKeyboardButtonRow(button))
	}

	keyboard := tgbotapi.NewReplyKeyboard(rows...)
	keyboard.ResizeKeyboard = true
	keyboard.OneTimeKeyboard = false

	return keyboard
}

func buildAssignedTaskKeyboard(dutyTasks []model.DutyTaskView, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, dutyTask := range dutyTasks {
		buttonText := fmt.Sprintf("%d этаж %s: %s", dutyTask.AreaFloor, dutyTask.AreaName, dutyTask.TaskTitle)
		callbackData := fmt.Sprintf("%s%s", callbackPrefix, dutyTask.TaskID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
