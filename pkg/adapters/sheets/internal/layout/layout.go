package layout

import (
	"dorm/pkg/adapters/sheets/utils"
	"dorm/pkg/core/ports/dto"
	"fmt"
)

func BuildSheetData(duty dto.DutyViewModel) [][]interface{} {
	data := [][]interface{}{
		{fmt.Sprintf("На этой неделе убирается команда %d", duty.TeamOrder)},
		{"Зона", "Задача", "Стоимость", "Исполнитель", "Статус"},
	}

	for _, task := range duty.Tasks {
		data = append(data, []interface{}{
			task.AreaName,
			task.Title,
			task.Cost,
			utils.FormatMemberName(task.Assignee, duty.UsersStats),
			utils.FormatStatus(task),
		})
	}

	return data
}

func PrepareStatsTable(duty dto.DutyViewModel) [][]interface{} {
	rows := [][]interface{}{
		{"Состав команды", "Бронь", "Заработано"},
	}
	for _, member := range duty.UsersStats {
		rows = append(rows, []interface{}{
			utils.FormatMemberName(member, duty.UsersStats),
			member.TotalPoints,
			member.ConfirmedPoints,
		})
	}
	rows = append(rows, getFooterInfo(duty)...)

	return rows
}

func getFooterInfo(duty dto.DutyViewModel) [][]interface{} {
	costSum := 0
	unassignedCount := 0
	for _, task := range duty.Tasks {
		costSum += task.Cost
		if task.Assignee == nil {
			unassignedCount++
		}
	}
	return [][]interface{}{
		{"Всего на члена команды:", fmt.Sprintf("%.1f", float64(costSum/len(duty.UsersStats)))},
		{"Задач не взято:", unassignedCount},
	}
}
