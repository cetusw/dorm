package telegram

import (
	"dorm/pkg/core/ports"
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	btnBack   = "« Назад"
	cbBack    = "back"
	cbAssign  = "cmd_assign"
	cbConfirm = "cmd_complete"
)

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🧹 Задачи"),
			tgbotapi.NewKeyboardButton("👤 Профиль"),
		),
	)
}

func taskMenuKeyboard() tgbotapi.InlineKeyboardMarkup {
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("➕ Взять задачу", cbAssign)),
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData("✅ Подтвердить выполнение", cbConfirm)),
	)
}

func areaSelectKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	sortTaskList(tasks)
	areaNames := make(map[int]string)
	for _, t := range tasks {
		areaNames[t.AreaID] = t.AreaName
	}

	ids := make([]int, 0, len(areaNames))
	for id := range areaNames {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, id := range ids {
		data := fmt.Sprintf("area:%d", id)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📍 "+areaNames[id], data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func taskSelectKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	sortTaskList(tasks)
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tasks {
		data := fmt.Sprintf("assign:%s", t.ID.String())
		text := fmt.Sprintf("%s (%d б.)", t.Title, t.Cost)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func taskConfirmKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	sortTaskList(tasks)
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tasks {
		data := fmt.Sprintf("complete:%s", t.ID.String())
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ "+t.Title, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func confirmAreaSelectKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	areaNames := make(map[int]string)
	for _, t := range tasks {
		areaNames[t.AreaID] = t.AreaName
	}

	ids := make([]int, 0, len(areaNames))
	for id := range areaNames {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, id := range ids {
		data := fmt.Sprintf("conf_area:%d", id)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ "+areaNames[id], data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func sortTaskList(tasks []ports.TaskViewModel) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Cost != tasks[j].Cost {
			return tasks[i].Cost > tasks[j].Cost
		}
		return tasks[i].Title < tasks[j].Title
	})
}
