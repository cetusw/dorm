package widgets

import (
	"fmt"
	"sort"

	"google.golang.org/api/sheets/v4"

	adapter "dorm/pkg/adapters/sheets"
	"dorm/pkg/adapters/sheets/domain"
	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/core/ports/dto"
)

const (
	userStatsWidth = 3
	statsHeaderRow = 1
	taskStartRow   = 3
)

type UserStatsWidget struct {
	duty      dto.DutyViewModel
	formatter *domain.SheetFormatter
}

func NewUserStatsWidget(duty dto.DutyViewModel) *UserStatsWidget {
	return &UserStatsWidget{duty: duty}
}

func (w *UserStatsWidget) GetWidth() int64 { return userStatsWidth }

func (w *UserStatsWidget) Render(sheetID int64, anchor types.Anchor) types.RenderResult {
	w.formatter = domain.NewSheetFormatter(sheetID)
	w.sortMembers()
	values := w.buildValues(anchor)

	var requests []*sheets.Request
	requests = append(requests, w.styleHeaders(anchor)...)
	requests = append(requests, w.styleBody(anchor)...)
	requests = append(requests, w.styleFooter(anchor)...)
	requests = append(requests, w.applyColumnWidths(anchor)...)

	return types.RenderResult{Values: values, Requests: requests}
}

func (w *UserStatsWidget) sortMembers() {
	sort.SliceStable(w.duty.UsersStats, func(i, j int) bool {
		return w.duty.UsersStats[i].IsTeamLeader && !w.duty.UsersStats[j].IsTeamLeader
	})
}

func (w *UserStatsWidget) buildValues(anchor types.Anchor) [][]interface{} {
	result := [][]interface{}{
		{""},
		{"Состав команды", "Бронь", "Заработано"},
	}

	costRng := w.getTaskColRef(2, len(w.duty.Tasks))
	assigneeRng := w.getTaskColRef(3, len(w.duty.Tasks))
	statusRng := w.getTaskColRef(4, len(w.duty.Tasks))

	for i, m := range w.duty.UsersStats {
		userNameCell := w.getA1(anchor.Col, int64(i+statsHeaderRow+1))
		result = append(result, []interface{}{
			adapter.FormatMemberName(m, w.duty.UsersStats),
			fmt.Sprintf("=SUMPRODUCT((%s=%s)*%s)", assigneeRng, userNameCell, costRng),
			fmt.Sprintf("=SUMPRODUCT((%s=%s)*%s*(%s=\"Проверено\"))", assigneeRng, userNameCell, costRng, statusRng),
		})
	}

	return append(result, w.buildFooterValues(assigneeRng, costRng)...)
}

func (w *UserStatsWidget) buildFooterValues(assigneeRng, costRng string) [][]interface{} {
	countMembers := len(w.duty.UsersStats)
	avgFormula := fmt.Sprintf("=SUM(%s)/%d", costRng, countMembers)
	unassignedFormula := fmt.Sprintf("=COUNTIF(%s; \"Никто\")", assigneeRng)

	return [][]interface{}{
		{"Всего на члена команды:", avgFormula, ""},
		{"Свободных задач:", unassignedFormula, ""},
	}
}

func (w *UserStatsWidget) getTaskColRef(colIdx, taskCount int) string {
	letter := string(rune('A' + colIdx))
	endRow := taskStartRow + taskCount - 1
	return fmt.Sprintf("$%s$%d:$%s$%d", letter, taskStartRow, letter, endRow)
}

func (w *UserStatsWidget) getA1(col, row int64) string {
	return fmt.Sprintf("%s%d", string(rune('A'+col)), row+1)
}

func (w *UserStatsWidget) styleHeaders(anchor types.Anchor) []*sheets.Request {
	row := anchor.Row + statsHeaderRow
	return []*sheets.Request{
		w.formatter.RepeatCell(w.formatter.NewRange(row, row+1, anchor.Col, anchor.Col+3), &sheets.CellFormat{
			BackgroundColor:     &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9},
			HorizontalAlignment: "CENTER",
			VerticalAlignment:   "MIDDLE",
			TextFormat:          &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(backgroundColor,horizontalAlignment,verticalAlignment,textFormat)"),
		w.formatter.UpdateBorders(w.formatter.NewRange(row, row+1, anchor.Col, anchor.Col+3), true, true, true, true, false, true),
	}
}

func (w *UserStatsWidget) styleBody(anchor types.Anchor) []*sheets.Request {
	userCount := len(w.duty.UsersStats)
	startRow := anchor.Row + statsHeaderRow + 1
	endRow := startRow + int64(userCount)
	teamColor, _ := adapter.HexToRGB(w.duty.TeamColor)

	requests := []*sheets.Request{
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, endRow, anchor.Col, anchor.Col+1), &sheets.CellFormat{
			BackgroundColor: teamColor, VerticalAlignment: "MIDDLE",
			TextFormat: &sheets.TextFormat{FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(backgroundColor,verticalAlignment,textFormat)"),
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, endRow, anchor.Col+1, anchor.Col+3), &sheets.CellFormat{
			HorizontalAlignment: "CENTER", VerticalAlignment: "MIDDLE",
			TextFormat: &sheets.TextFormat{FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(horizontalAlignment,verticalAlignment,textFormat)"),
		w.formatter.UpdateBorders(w.formatter.NewRange(startRow, endRow, anchor.Col, anchor.Col+3), true, true, true, true, false, true),
	}

	if userCount > 0 {
		requests = append(requests, w.styleCaptainRow(startRow, anchor.Col))
	}

	return requests
}

func (w *UserStatsWidget) styleCaptainRow(row, col int64) *sheets.Request {
	return w.formatter.RepeatCell(w.formatter.NewRange(row, row+1, col, col+3), &sheets.CellFormat{
		TextFormat: &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
	}, "userEnteredFormat.textFormat")
}

func (w *UserStatsWidget) styleFooter(anchor types.Anchor) []*sheets.Request {
	userCount := int64(len(w.duty.UsersStats))
	startRow := anchor.Row + statsHeaderRow + 1 + userCount

	reqs := []*sheets.Request{
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, startRow+2, anchor.Col, anchor.Col+3), &sheets.CellFormat{
			VerticalAlignment: "MIDDLE",
			TextFormat:        &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
		}, "userEnteredFormat(verticalAlignment,textFormat)"),
		w.formatter.MergeCells(startRow, startRow+1, anchor.Col+1, anchor.Col+3, "MERGE_ALL"),
		w.formatter.MergeCells(startRow+1, startRow+2, anchor.Col+1, anchor.Col+3, "MERGE_ALL"),
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, startRow+2, anchor.Col, anchor.Col+1), &sheets.CellFormat{HorizontalAlignment: "RIGHT"}, "userEnteredFormat.horizontalAlignment"),
		w.formatter.RepeatCell(w.formatter.NewRange(startRow, startRow+2, anchor.Col+1, anchor.Col+2), &sheets.CellFormat{HorizontalAlignment: "LEFT"}, "userEnteredFormat.horizontalAlignment"),
		w.formatter.UpdateBorders(w.formatter.NewRange(startRow, startRow+2, anchor.Col, anchor.Col+3), true, true, true, true, true, false),
	}

	return append(reqs, w.addFooterConditionalFormatting(startRow, anchor.Col)...)
}

func (w *UserStatsWidget) addFooterConditionalFormatting(row, col int64) []*sheets.Request {
	columnLetter := string(rune('A' + col + 1))
	valCellAbs := fmt.Sprintf("$%s$%d", columnLetter, row+2)

	startCol := col
	endCol := col + 3

	return []*sheets.Request{
		w.createBooleanRule(row+1, row+2, startCol, endCol, fmt.Sprintf("=%s>0", valCellAbs), "fde0a3"),
		w.createBooleanRule(row+1, row+2, startCol, endCol, fmt.Sprintf("=%s=0", valCellAbs), "d0eab7"),
	}
}

func (w *UserStatsWidget) createBooleanRule(rs, re, cs, ce int64, formula, color string) *sheets.Request {
	bg, _ := adapter.HexToRGB(color)
	return &sheets.Request{
		AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
			Rule: &sheets.ConditionalFormatRule{
				Ranges: []*sheets.GridRange{w.formatter.NewRange(rs, re, cs, ce)},
				BooleanRule: &sheets.BooleanRule{
					Condition: &sheets.BooleanCondition{Type: "CUSTOM_FORMULA", Values: []*sheets.ConditionValue{{UserEnteredValue: formula}}},
					Format:    &sheets.CellFormat{BackgroundColor: bg},
				},
			},
		},
	}
}

func (w *UserStatsWidget) applyColumnWidths(anchor types.Anchor) []*sheets.Request {
	widths := []int64{190, 100, 100}
	var requests []*sheets.Request
	for i, width := range widths {
		requests = append(requests, w.formatter.SetDimensionSize("COLUMNS", anchor.Col+int64(i), width))
	}
	return requests
}
