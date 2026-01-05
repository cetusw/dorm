package telegram

import (
	"dorm/pkg/core/ports"
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	cbBack    = "back"
	cbAssign  = "cmd_assign"
	cbConfirm = "cmd_complete"
)

func mainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnProfile),
			tgbotapi.NewKeyboardButton(btnDuty),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton(btnMyTasks),
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
	areas := getSortedAreas(tasks)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, a := range areas {
		data := fmt.Sprintf("area:%d", a.id)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(a.name, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func taskSelectKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	SortTasks(tasks)
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
	SortTasks(tasks)
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, t := range tasks {
		data := fmt.Sprintf("complete:%s", t.ID.String())
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(t.Title, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func confirmAreaSelectKeyboard(tasks []ports.TaskViewModel) tgbotapi.InlineKeyboardMarkup {
	areas := getSortedAreas(tasks)

	var rows [][]tgbotapi.InlineKeyboardButton
	for _, a := range areas {
		data := fmt.Sprintf("conf_area:%d", a.id)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(a.name, data),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(btnBack, cbBack)))
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

func SortTasks(tasks []ports.TaskViewModel) {
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].Cost != tasks[j].Cost {
			return tasks[i].Cost > tasks[j].Cost
		}
		return tasks[i].Title < tasks[j].Title
	})
}

type area struct {
	id    int
	name  string
	floor int
}

type areaList []area

func (a areaList) Len() int           { return len(a) }
func (a areaList) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a areaList) Less(i, j int) bool { return a[i].name < a[j].name }

func getSortedAreas(tasks []ports.TaskViewModel) areaList {
	areaMap := make(map[int]area)
	for _, t := range tasks {
		if _, ok := areaMap[t.AreaID]; !ok {
			areaMap[t.AreaID] = area{
				id:    t.AreaID,
				name:  fmt.Sprintf("%d этаж. %s", t.AreaFloor, t.AreaName),
				floor: t.AreaFloor,
			}
		}
	}

	areas := make(areaList, 0, len(areaMap))
	for _, a := range areaMap {
		areas = append(areas, a)
	}
	sort.Sort(areas)
	return areas
}
