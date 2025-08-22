package keyboard

import (
	"dorm/internal/common/message"
	"dorm/internal/dorm/application/model"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	CallbackPrefixArea     = "area_select_"
	CallbackPrefixTask     = "task_select_"
	CallbackPrefixConfirm  = "confirm_task_"
	CallbackPrefixUnassign = "unassign_task_"
)

func BuildMainStateKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return buildReplyKeyboard(
		[]buttonInfo{
			{message.Tasks, message.Tasks},
			{message.Team, message.Team},
			{message.Payment, message.Payment},
			{message.Profile, message.Profile},
		},
	)
}

func BuildTaskManagementKeyboard() tgbotapi.InlineKeyboardMarkup {
	return buildSimpleKeyboard(
		[]buttonInfo{
			{message.ConfirmExecution, message.ConfirmExecution},
			{message.AssignTask, message.AssignTask},
			{message.UnassignTask, message.UnassignTask},
			{message.Back, message.Back},
		},
	)
}

func BuildTeamManagementKeyboard() tgbotapi.InlineKeyboardMarkup {
	return buildSimpleKeyboard(
		[]buttonInfo{
			{message.TeamMembers, message.TeamMembers},
			{message.DutySchedule, message.DutySchedule},
			{message.Back, message.Back},
		},
	)
}

func BuildBackKeyboard() tgbotapi.InlineKeyboardMarkup {
	return buildSimpleKeyboard([]buttonInfo{{message.Back, message.Back}})
}

func BuildAreaKeyboard(areas []model.Area) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, area := range areas {
		callbackData := fmt.Sprintf("%s%d", CallbackPrefixArea, area.AreaID)
		button := tgbotapi.NewInlineKeyboardButtonData(area.Name, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildTaskKeyboard(tasks []model.Task) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, task := range tasks {
		callbackData := fmt.Sprintf("%s%s", CallbackPrefixTask, task.TaskID.String())
		button := tgbotapi.NewInlineKeyboardButtonData(task.Title, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func BuildConfirmExecutionKeyboard(assignedTasks []model.DutyTask) tgbotapi.InlineKeyboardMarkup {
	return buildAssignedTaskKeyboard(assignedTasks, CallbackPrefixConfirm)
}

func BuildTaskUnassignmentKeyboard(assignedTasks []model.DutyTask) tgbotapi.InlineKeyboardMarkup {
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
	// TODO: изучить, что такое keyboard.OneTimeKeyboard
	// OneTimeKeyboard можно убрать для главного меню, чтобы оно не скрывалось
	// keyboard.OneTimeKeyboard = true

	return keyboard
}

func buildAssignedTaskKeyboard(dutyTasks []model.DutyTask, callbackPrefix string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, dutyTask := range dutyTasks {
		buttonText := fmt.Sprintf("%s: %s", dutyTask.TaskID, dutyTask.TaskID) // TODO: придумать, как получить название задачи из dutyTask. Как-то нужно заюзать GetTaskTitleByTaskID
		callbackData := fmt.Sprintf("%s%s", callbackPrefix, dutyTask.TaskID)
		button := tgbotapi.NewInlineKeyboardButtonData(buttonText, callbackData)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(button))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(message.Back, message.Back)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}
