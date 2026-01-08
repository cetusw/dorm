package layout

import (
	"dorm/pkg/core/ports/dto"
	"fmt"
)

func BuildSheetData(duty dto.DutyViewModel) [][]interface{} {
	header := []interface{}{"Зона", "Задача", "Баллы", "Исполнитель", "Статус"}
	data := [][]interface{}{
		{fmt.Sprintf("Команда: %d", duty.TeamOrder)},
		{},
		header,
	}

	for _, t := range duty.Tasks {
		state := "Не сделано"
		if t.IsDone {
			state = "Сделано"
		}
		row := []interface{}{
			t.AreaName,
			t.Title,
			t.Cost,
			t.Executor,
			state,
		}
		data = append(data, row)
	}

	return data
}
