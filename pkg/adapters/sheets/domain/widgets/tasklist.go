package widgets

import (
	"fmt"

	"google.golang.org/api/sheets/v4"

	adapter "dorm/pkg/adapters/sheets"
	"dorm/pkg/adapters/sheets/domain"
	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/core/ports/dto"
)

const (
	taskListWidth = 5
	headerRows    = 2
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

	tasks := adapter.SortForTableView(w.duty.Tasks)
	values := w.buildValues(tasks)

	var requests []*sheets.Request
	requests = append(requests, w.styleMainHeader(anchor)...)
	requests = append(requests, w.styleTableColumns(anchor)...)
	requests = append(requests, w.styleTableBody(anchor, len(tasks))...)
	requests = append(requests, w.mergeAreas(anchor, tasks)...)
	requests = append(requests, w.styleAreas(anchor, tasks)...)
	requests = append(requests, w.addCostGradient(anchor, len(tasks))...)
	requests = append(requests, w.addValidations(anchor, len(tasks))...)
	requests = append(requests, w.applyColumnWidths(anchor)...)

	return types.RenderResult{
		Values:   values,
		Requests: requests,
	}
}

func (w *TaskListWidget) buildValues(tasks []dto.TaskViewModel) [][]interface{} {
	res := [][]interface{}{
		{fmt.Sprintf("На этой неделе убирается команда %d", w.duty.TeamOrder)},
		{"Зона", "Задача", "Стоимость", "Исполнитель", "Статус"},
	}

	for _, task := range tasks {
		res = append(res, []interface{}{
			fmt.Sprintf("%d этаж. %s", task.AreaFloor, task.AreaName),
			task.Title,
			task.Cost,
			adapter.FormatMemberName(task.Assignee, w.duty.UsersStats),
			adapter.FormatStatus(task),
		})
	}
	return res
}

func (w *TaskListWidget) styleMainHeader(anchor types.Anchor) []*sheets.Request {
	teamColor, _ := adapter.HexToRGB(w.duty.TeamColor)
	darkenedText := adapter.DarkenColor(teamColor)

	return []*sheets.Request{
		w.formatter.MergeCells(anchor.Row, anchor.Row+1, anchor.Col, anchor.Col+5, "MERGE_ALL"),
		w.formatter.RepeatCell(w.formatter.NewRange(anchor.Row, anchor.Row+1, anchor.Col, anchor.Col+5), &sheets.CellFormat{
			BackgroundColor:     teamColor,
			HorizontalAlignment: "CENTER",
			VerticalAlignment:   "MIDDLE",
			TextFormat:          &sheets.TextFormat{Bold: true, FontSize: 14, FontFamily: "Montserrat", ForegroundColor: darkenedText},
		}, "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"),
		w.formatter.SetDimensionSize("ROWS", anchor.Row, 50),
	}
}

func (w *TaskListWidget) styleTableColumns(anchor types.Anchor) []*sheets.Request {
	grayBg := &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9}
	row := anchor.Row + 1

	return []*sheets.Request{
		w.formatter.RepeatCell(w.formatter.NewRange(row, row+1, anchor.Col, anchor.Col+5), &sheets.CellFormat{
			BackgroundColor:     grayBg,
			HorizontalAlignment: "CENTER",
			VerticalAlignment:   "MIDDLE",
			TextFormat:          &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"),

		w.formatter.UpdateBorders(w.formatter.NewRange(row, row+1, anchor.Col, anchor.Col+5), true, true, true, true, false, true),
	}
}

func (w *TaskListWidget) styleTableBody(anchor types.Anchor, taskCount int) []*sheets.Request {
	startRow := anchor.Row + headerRows
	endRow := startRow + int64(taskCount)
	bodyColor := &sheets.Color{Red: 0.95, Green: 0.95, Blue: 0.95} // #f3f3f3

	return []*sheets.Request{
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, endRow, anchor.Col, anchor.Col+5), &sheets.CellFormat{
			BackgroundColor:   bodyColor,
			VerticalAlignment: "MIDDLE",
			TextFormat:        &sheets.TextFormat{FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(backgroundColor,textFormat,verticalAlignment)"),

		w.formatter.RepeatCell(w.formatter.NewRange(startRow, endRow, anchor.Col+1, anchor.Col+2), &sheets.CellFormat{HorizontalAlignment: "LEFT"}, "userEnteredFormat.horizontalAlignment"),
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, endRow, anchor.Col+2, anchor.Col+5), &sheets.CellFormat{HorizontalAlignment: "CENTER"}, "userEnteredFormat.horizontalAlignment"),
	}
}

func (w *TaskListWidget) getAreaRowRanges(anchor types.Anchor, tasks []dto.TaskViewModel) []adapter.RowRange {
	var ranges []adapter.RowRange
	if len(tasks) == 0 {
		return ranges
	}

	startRow := anchor.Row + headerRows
	blockStart := startRow

	for i := 1; i <= len(tasks); i++ {
		isLast := i == len(tasks)
		var areaChanged bool
		if !isLast {
			areaChanged = tasks[i].AreaName != tasks[i-1].AreaName || tasks[i].AreaFloor != tasks[i-1].AreaFloor
		}

		if areaChanged || isLast {
			blockEnd := startRow + int64(i)
			ranges = append(ranges, adapter.RowRange{Start: blockStart, End: blockEnd})
			blockStart = blockEnd
		}
	}
	return ranges
}

func (w *TaskListWidget) mergeAreas(anchor types.Anchor, tasks []dto.TaskViewModel) []*sheets.Request {
	ranges := w.getAreaRowRanges(anchor, tasks)
	var reqs []*sheets.Request

	for _, r := range ranges {
		if r.End-r.Start > 1 {
			reqs = append(reqs, w.formatter.MergeCells(r.Start, r.End, anchor.Col, anchor.Col+1, "MERGE_ALL"))
		}
	}
	return reqs
}

func (w *TaskListWidget) styleAreas(anchor types.Anchor, tasks []dto.TaskViewModel) []*sheets.Request {
	ranges := w.getAreaRowRanges(anchor, tasks)
	var requests []*sheets.Request

	for _, r := range ranges {
		requests = append(requests, w.formatter.RepeatCell(w.formatter.NewRange(r.Start, r.End, anchor.Col, anchor.Col+1), &sheets.CellFormat{
			VerticalAlignment:   "MIDDLE",
			HorizontalAlignment: "CENTER",
			WrapStrategy:        "WRAP",
			TextFormat:          &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(verticalAlignment,horizontalAlignment,wrapStrategy,textFormat)"))

		requests = append(requests, w.formatter.UpdateBorders(
			w.formatter.NewRange(r.Start, r.End, anchor.Col, anchor.Col+5),
			true,
			true,
			true,
			true,
			false,
			true,
		))
	}
	return requests
}

func (w *TaskListWidget) addCostGradient(anchor types.Anchor, taskCount int) []*sheets.Request {
	minPoint, _ := adapter.HexToRGB("d0eab7")
	midPoint, _ := adapter.HexToRGB("fde0a3")
	maxPoint, _ := adapter.HexToRGB("f4c1c1")

	return []*sheets.Request{{
		AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
			Rule: &sheets.ConditionalFormatRule{
				Ranges: []*sheets.GridRange{w.formatter.NewRange(anchor.Row+headerRows, anchor.Row+headerRows+int64(taskCount), anchor.Col+2, anchor.Col+3)},
				GradientRule: &sheets.GradientRule{
					Minpoint: &sheets.InterpolationPoint{Color: minPoint, Type: "NUMBER", Value: "1"},
					Midpoint: &sheets.InterpolationPoint{Color: midPoint, Type: "NUMBER", Value: "5"},
					Maxpoint: &sheets.InterpolationPoint{Color: maxPoint, Type: "NUMBER", Value: "9"},
				},
			},
		},
	}}
}

func (w *TaskListWidget) addValidations(anchor types.Anchor, taskCount int) []*sheets.Request {
	startRow := anchor.Row + headerRows
	endRow := startRow + int64(taskCount)

	var memberValues []*sheets.ConditionValue
	for _, m := range w.duty.UsersStats {
		memberValues = append(memberValues, &sheets.ConditionValue{UserEnteredValue: adapter.FormatMemberName(m, w.duty.UsersStats)})
	}
	memberValues = append(memberValues, &sheets.ConditionValue{UserEnteredValue: "Никто"})

	return []*sheets.Request{
		{
			SetDataValidation: &sheets.SetDataValidationRequest{
				Range: w.formatter.NewRange(startRow, endRow, anchor.Col+3, anchor.Col+4),
				Rule:  &sheets.DataValidationRule{Condition: &sheets.BooleanCondition{Type: "ONE_OF_LIST", Values: memberValues}, ShowCustomUi: true},
			},
		},
		{
			RepeatCell: &sheets.RepeatCellRequest{
				Range: w.formatter.NewRange(startRow, endRow, anchor.Col+4, anchor.Col+5),
				Cell: &sheets.CellData{
					DataValidation: &sheets.DataValidationRule{
						Condition: &sheets.BooleanCondition{
							Type:   "ONE_OF_LIST",
							Values: []*sheets.ConditionValue{{UserEnteredValue: "Не сделано"}, {UserEnteredValue: "Сделано"}, {UserEnteredValue: "Проверено"}},
						},
						ShowCustomUi: true,
					},
				},
				Fields: "dataValidation",
			},
		},
	}
}

func (w *TaskListWidget) applyColumnWidths(anchor types.Anchor) []*sheets.Request {
	widths := []int64{100, 380, 80, 120, 120}
	var requests []*sheets.Request
	for i, width := range widths {
		requests = append(requests, w.formatter.SetDimensionSize("COLUMNS", anchor.Col+int64(i), width))
	}
	return requests
}
