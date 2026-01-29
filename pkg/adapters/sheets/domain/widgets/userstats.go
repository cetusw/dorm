package widgets

import (
	sheets2 "dorm/pkg/adapters/sheets"
	"dorm/pkg/adapters/sheets/domain"

	"google.golang.org/api/sheets/v4"

	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/core/ports/dto"
)

const (
	userStatsWidth = 3
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
	result := types.RenderResult{}

	actualRow := anchor.Row + 1

	result.Values = [][]interface{}{
		{"Состав команды", "Бронь", "Заработано"},
	}
	for _, m := range w.duty.UsersStats {
		result.Values = append(result.Values, []interface{}{
			sheets2.FormatMemberName(m, w.duty.UsersStats), m.TotalPoints, m.ConfirmedPoints,
		})
	}

	userColor, _ := sheets2.HexToRGB(w.duty.TeamColor)
	endRow := actualRow + int64(len(w.duty.UsersStats)) + 1

	result.Requests = append(result.Requests,
		w.formatter.RepeatCell(w.formatter.NewRange(actualRow, actualRow+1, anchor.Col, anchor.Col+3), &sheets.CellFormat{
			BackgroundColor:     &sheets.Color{Red: 0.9, Green: 0.9, Blue: 0.9},
			TextFormat:          &sheets.TextFormat{Bold: true, FontSize: 10, FontFamily: "Montserrat"},
			HorizontalAlignment: "CENTER", VerticalAlignment: "MIDDLE",
		}, "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment,verticalAlignment)"),
		w.formatter.RepeatCell(w.formatter.NewRange(actualRow+1, endRow, anchor.Col, anchor.Col+1), &sheets.CellFormat{
			BackgroundColor: userColor, VerticalAlignment: "MIDDLE",
		}, "userEnteredFormat(backgroundColor,verticalAlignment)"),
	)

	return result
}
