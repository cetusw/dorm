package sheets

import (
	"google.golang.org/api/sheets/v4"

	"dorm/pkg/core/ports/dto"
)

const (
	headerColor = "E0E0E0"
	borderColor = "000000"
)

func ApplyHeaderStyle(sheetID int64) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    2,
				EndRowIndex:      3,
				StartColumnIndex: 0,
				EndColumnIndex:   5,
			},
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor: &sheets.Color{
						Red:   0.878,
						Green: 0.878,
						Blue:  0.878,
					},
					TextFormat: &sheets.TextFormat{
						Bold: true,
					},
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat)",
		},
	}
}

func ApplyTaskBorders(sheetID int64, rowCount, colCount int) *sheets.Request {
	return &sheets.Request{
		UpdateBorders: &sheets.UpdateBordersRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    2,
				EndRowIndex:      int64(2 + rowCount + 1),
				StartColumnIndex: 0,
				EndColumnIndex:   int64(colCount),
			},
			Top: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
			Bottom: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
			Left: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
			Right: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
			InnerHorizontal: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
			InnerVertical: &sheets.Border{
				Style: "SOLID",
				Color: &sheets.Color{Red: 0, Green: 0, Blue: 0},
			},
		},
	}
}

func MergeAreaCells(sheetID int64, tasks []dto.TaskViewModel) []*sheets.Request {
	var requests []*sheets.Request
	if len(tasks) <= 1 {
		return requests
	}

	startMergeRowIndex := 0
	for i := 1; i < len(tasks); i++ {
		if tasks[i].AreaName != tasks[startMergeRowIndex].AreaName {
			if i-startMergeRowIndex > 1 {
				requests = append(requests, &sheets.Request{
					MergeCells: &sheets.MergeCellsRequest{
						Range: &sheets.GridRange{
							SheetId:          sheetID,
							StartRowIndex:    int64(startMergeRowIndex + 3),
							EndRowIndex:      int64(i + 3),
							StartColumnIndex: 0,
							EndColumnIndex:   1,
						},
						MergeType: "MERGE_ROWS",
					},
				})
			}
			startMergeRowIndex = i
		}
	}

	if len(tasks)-startMergeRowIndex > 1 {
		requests = append(requests, &sheets.Request{
			MergeCells: &sheets.MergeCellsRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    int64(startMergeRowIndex + 3),
					EndRowIndex:      int64(len(tasks) + 3),
					StartColumnIndex: 0,
					EndColumnIndex:   1,
				},
				MergeType: "MERGE_ROWS",
			},
		})
	}

	return requests
}

func AddStatusValidation(sheetID int64, rowCount int) *sheets.Request {
	return &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    3,
				EndRowIndex:      int64(3 + rowCount),
				StartColumnIndex: 4,
				EndColumnIndex:   5,
			},
			Cell: &sheets.CellData{
				DataValidation: &sheets.DataValidationRule{
					Condition: &sheets.BooleanCondition{
						Type:   "ONE_OF_LIST",
						Values: []*sheets.ConditionValue{{UserEnteredValue: "Не сделано"}, {UserEnteredValue: "Сделано"}},
					},
					ShowCustomUi: true,
				},
			},
			Fields: "dataValidation",
		},
	}
}
