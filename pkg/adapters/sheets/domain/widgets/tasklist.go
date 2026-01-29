package widgets

import (
	sheets2 "dorm/pkg/adapters/sheets"
	"dorm/pkg/adapters/sheets/domain"
	"fmt"

	"google.golang.org/api/sheets/v4"

	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/core/ports/dto"
)

const (
	taskListWidth = 6
)

type TaskListWidget struct {
	duty      dto.DutyViewModel
	formatter *domain.SheetFormatter
}

func NewTaskListWidget(duty dto.DutyViewModel) *TaskListWidget {
	return &TaskListWidget{duty: duty}
}

func (w *TaskListWidget) GetWidth() int64 { return taskListWidth }

func (w *TaskListWidget) Render(sheetID int64, anchor types.Anchor) types.RenderResult {
	w.formatter = domain.NewSheetFormatter(sheetID)
	result := types.RenderResult{}

	result.Values = [][]interface{}{
		{fmt.Sprintf("На этой неделе убирается команда %d", w.duty.TeamOrder)},
		{"Зона", "Задача", "Стоимость", "Исполнитель", "Статус"},
	}
	sortedTasks := sheets2.SortForTableView(w.duty.Tasks)
	for _, task := range sortedTasks {
		result.Values = append(result.Values, []interface{}{
			fmt.Sprintf("%d этаж. %s", task.AreaFloor, task.AreaName),
			task.Title,
			task.Cost,
			sheets2.FormatMemberName(task.Assignee, w.duty.UsersStats),
			sheets2.FormatStatus(task),
		})
	}

	teamColor, _ := sheets2.HexToRGB(w.duty.TeamColor)
	darkenedText := sheets2.DarkenColor(teamColor)

	result.Requests = append(result.Requests,
		w.formatter.MergeCells(anchor.Row, anchor.Row+1, anchor.Col, anchor.Col+5, "MERGE_ALL"),
		w.formatter.RepeatCell(w.formatter.NewRange(anchor.Row, anchor.Row+1, anchor.Col, anchor.Col+5), &sheets.CellFormat{
			BackgroundColor: teamColor, HorizontalAlignment: "CENTER", VerticalAlignment: "MIDDLE",
			TextFormat: &sheets.TextFormat{Bold: true, FontSize: 14, FontFamily: "Montserrat", ForegroundColor: darkenedText},
		}, "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"),
	)

	return result
}
