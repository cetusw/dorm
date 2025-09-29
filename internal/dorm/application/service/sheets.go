package service

import (
	"dorm/internal/common/consts"
	"dorm/internal/common/utils"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

type SheetsService struct {
	sheets *infrastructure.Sheets
}

func NewSheetsService(sheets *infrastructure.Sheets) *SheetsService {
	return &SheetsService{
		sheets: sheets,
	}
}

func (s *SheetsService) CreateDutySheet(sheetData model.SheetData) error {
	teamColor, err := utils.HexToSheetsColor(sheetData.TeamColor)
	if err != nil {
		return err
	}
	sheet, err := s.sheets.CreateSheet(sheetData.Title, teamColor, sheetData.SpreadsheetID)
	if err != nil {
		return err
	}
	sheetID := sheet.Properties.SheetId

	dataToWrite, zoneMergeRanges := s.prepareSheetData(sheetData.Order, sheetData.Tasks)

	err = s.sheets.WriteRange(sheetData, "A1", dataToWrite)
	if err != nil {
		return err
	}

	userTableData := s.prepareUserTableData(sheetData.Users)
	if len(userTableData) > 0 {
		err = s.sheets.WriteRange(sheetData, "G2", userTableData)
		if err != nil {
			return fmt.Errorf("failed to write user table to column G: %w", err)
		}

		requests := s.createUserTableFormattingRequests(sheetID, len(userTableData), teamColor)
		err = s.sheets.BatchUpdate(sheetData, requests)
		if err != nil {
			return fmt.Errorf("failed to format user table: %w", err)
		}
	}

	requests := s.prepareFormattingRequests(sheetID, zoneMergeRanges, teamColor, len(sheetData.Tasks), sheetData.Users)
	return s.sheets.BatchUpdate(sheetData, requests)
}

func (s *SheetsService) UpdateDutySheet(sheetData model.SheetData) error {
	var dataToWrite [][]interface{}
	for _, task := range sheetData.Tasks {
		row := s.formatTaskRow(task)
		dataToWrite = append(dataToWrite, row)
	}

	clearRange := fmt.Sprintf("A%d:K", consts.TasksStartRow)
	err := s.sheets.ClearRange(sheetData, clearRange)
	if err != nil {
		return fmt.Errorf("failed to clear range in sheet '%s': %w", sheetData.Title, err)
	}

	writeRange := fmt.Sprintf("A%d", consts.TasksStartRow)
	err = s.sheets.WriteRange(sheetData, writeRange, dataToWrite)
	if err != nil {
		return fmt.Errorf("failed to write new data to sheet '%s': %w", sheetData.Title, err)
	}

	return nil
}

func (s *SheetsService) prepareUserTableData(users []model.User) [][]interface{} {
	var data [][]interface{}

	data = append(data, []interface{}{
		"Состав команды",
		"Бронь",
		"Заработано",
	})

	for i, user := range users {
		name := fmt.Sprintf("%s %s.", user.FirstName, string([]rune(user.LastName)[0]))
		data = append(data, []interface{}{
			name,
			fmt.Sprintf("=SUMPRODUCT(($D$3:$D = G%d)*$C$3:$C)", i+3),
			fmt.Sprintf("=SUMPRODUCT(($D$3:$D = G%d)*$C$3:$C*($E$3:$E = \"Проверено\"))", i+3),
		})
	}

	data = append(data, []interface{}{
		"",
		"",
		"",
	})

	return data
}

func (s *SheetsService) createUserTableFormattingRequests(sheetID int64, rowCount int, teamColor *sheets.Color) []*sheets.Request {
	var requests []*sheets.Request

	headerRange := &sheets.GridRange{
		SheetId:          sheetID,
		StartRowIndex:    1,
		EndRowIndex:      2,
		StartColumnIndex: 6,
		EndColumnIndex:   9,
	}
	requests = append(requests, &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: headerRange,
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor:     consts.SubHeaderBgColor,
					HorizontalAlignment: "CENTER",
					TextFormat:          &sheets.TextFormat{Bold: true},
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	})

	dataRange := &sheets.GridRange{
		SheetId:          sheetID,
		StartRowIndex:    2,
		EndRowIndex:      int64(rowCount),
		StartColumnIndex: 6,
		EndColumnIndex:   9,
	}
	requests = append(requests, &sheets.Request{
		RepeatCell: &sheets.RepeatCellRequest{
			Range: dataRange,
			Cell: &sheets.CellData{
				UserEnteredFormat: &sheets.CellFormat{
					BackgroundColor:     teamColor,
					HorizontalAlignment: "LEFT",
					TextFormat:          &sheets.TextFormat{FontSize: 12},
				},
			},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		},
	})

	columnWidths := []int{150, 80, 80}
	for i, width := range columnWidths {
		colStart := int64(6 + i)
		requests = append(requests, &sheets.Request{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: colStart,
					EndIndex:   colStart + 1,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: int64(width),
				},
				Fields: "pixelSize",
			},
		})
	}

	return requests
}

func (s *SheetsService) prepareSheetData(
	order int,
	tasks []model.DutyTaskView,
) ([][]interface{}, map[string][2]int) {
	var dataToWrite [][]interface{}
	dataToWrite = append(dataToWrite, []interface{}{fmt.Sprintf(consts.SheetHeaderTeam, order)})
	dataToWrite = append(dataToWrite, []interface{}{
		consts.SheetColumnArea,
		consts.SheetColumnTask,
		consts.SheetColumnCost,
		consts.SheetColumnAssignee,
		consts.SheetColumnState,
	})

	zoneMergeRanges := make(map[string][2]int)

	for i, task := range tasks {
		row := s.formatTaskRow(task)
		dataToWrite = append(dataToWrite, row)

		areaFullName := fmt.Sprintf(consts.SheetRowArea, task.AreaFloor, task.AreaName)
		if ranges, ok := zoneMergeRanges[areaFullName]; ok {
			ranges[1] = consts.TasksStartRow + i
			zoneMergeRanges[areaFullName] = ranges
		} else {
			zoneMergeRanges[areaFullName] = [2]int{consts.TasksStartRow + i, consts.TasksStartRow + i}
		}
	}
	return dataToWrite, zoneMergeRanges
}

func (s *SheetsService) formatTaskRow(task model.DutyTaskView) []interface{} {
	assignee := consts.DefaultAssignee
	if task.AssigneeFirstName != nil && task.AssigneeLastName != nil {
		assignee = fmt.Sprintf("%s %s.", *task.AssigneeFirstName, string([]rune(*task.AssigneeLastName)[0]))
	}

	state := consts.StateNotDone
	if task.CompletionDate != nil {
		state = consts.StateDone
	}
	if task.VerificationDate != nil {
		state = consts.StateVerified
	}

	return []interface{}{
		fmt.Sprintf("%d этаж. %s", task.AreaFloor, task.AreaName),
		task.TaskTitle,
		task.TaskCost,
		assignee,
		state,
	}
}

func (s *SheetsService) prepareFormattingRequests(sheetID int64, zoneMergeRanges map[string][2]int, teamColor *sheets.Color, rows int, users []model.User) []*sheets.Request {
	var requests []*sheets.Request

	requests = append(requests, s.createMergeRequests(sheetID, zoneMergeRanges)...)
	requests = append(requests, s.createStyleRequests(sheetID, teamColor)...)
	requests = append(requests, s.createValidationRequests(sheetID, rows, users)...)
	requests = append(requests, s.createConditionalFormattingRequests(sheetID)...)
	requests = append(requests, s.createDimensionRequests(sheetID)...)

	return requests
}

func (s *SheetsService) createMergeRequests(sheetID int64, zoneMergeRanges map[string][2]int) []*sheets.Request {
	requests := []*sheets.Request{
		{MergeCells: &sheets.MergeCellsRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 0,
				EndColumnIndex:   5,
			},
			MergeType: "MERGE_ALL",
		}},
	}
	for _, ranges := range zoneMergeRanges {
		if ranges[0] >= ranges[1] {
			continue
		}
		requests = append(requests, &sheets.Request{
			MergeCells: &sheets.MergeCellsRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    int64(ranges[0] - 1),
					EndRowIndex:      int64(ranges[1]),
					StartColumnIndex: 0,
					EndColumnIndex:   1,
				},
				MergeType: "MERGE_COLUMNS",
			},
		})
	}
	return requests
}

func (s *SheetsService) createStyleRequests(sheetID int64, teamColor *sheets.Color) []*sheets.Request {
	return []*sheets.Request{
		{RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    0,
				EndRowIndex:      1,
				StartColumnIndex: 0,
				EndColumnIndex:   5,
			},
			Cell: &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{
				BackgroundColor:     teamColor,
				HorizontalAlignment: "CENTER",
				VerticalAlignment:   "MIDDLE",
				TextFormat:          &sheets.TextFormat{FontSize: 14, Bold: true},
			}},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		}},
		{RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    1,
				EndRowIndex:      2,
				StartColumnIndex: 0,
				EndColumnIndex:   5,
			},
			Cell: &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{
				BackgroundColor:     consts.SubHeaderBgColor,
				HorizontalAlignment: "CENTER",
				TextFormat:          &sheets.TextFormat{Bold: true},
			}},
			Fields: "userEnteredFormat(backgroundColor,textFormat,horizontalAlignment)",
		}},
		{RepeatCell: &sheets.RepeatCellRequest{
			Range: &sheets.GridRange{
				SheetId:          sheetID,
				StartRowIndex:    int64(consts.TasksStartRow - 1),
				StartColumnIndex: 0,
				EndColumnIndex:   1,
			},
			Cell: &sheets.CellData{UserEnteredFormat: &sheets.CellFormat{
				VerticalAlignment:   "MIDDLE",
				HorizontalAlignment: "CENTER",
				WrapStrategy:        "WRAP",
				TextFormat:          &sheets.TextFormat{Bold: true},
				TextRotation:        &sheets.TextRotation{Angle: 90},
			}},
			Fields: "userEnteredFormat(verticalAlignment,horizontalAlignment,wrapStrategy,textFormat,textRotation)",
		}},
	}
}

func (s *SheetsService) createValidationRequests(sheetID int64, rows int, users []model.User) []*sheets.Request {
	var userConditionValues []*sheets.ConditionValue
	userConditionValues = append(userConditionValues, &sheets.ConditionValue{UserEnteredValue: consts.DefaultAssignee})
	for _, user := range users {
		userName := fmt.Sprintf("%s %s.", user.FirstName, string([]rune(user.LastName)[0]))
		userConditionValues = append(userConditionValues, &sheets.ConditionValue{UserEnteredValue: userName})
	}

	stateConditionValues := []*sheets.ConditionValue{
		{UserEnteredValue: consts.StateNotDone},
		{UserEnteredValue: consts.StateDone},
		{UserEnteredValue: consts.StateVerified},
	}

	return []*sheets.Request{
		{
			SetDataValidation: &sheets.SetDataValidationRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    int64(consts.TasksStartRow - 1),
					EndRowIndex:      int64(consts.TasksStartRow + rows - 1),
					StartColumnIndex: 3,
					EndColumnIndex:   4,
				},
				Rule: &sheets.DataValidationRule{
					Condition: &sheets.BooleanCondition{
						Type:   "ONE_OF_LIST",
						Values: userConditionValues,
					},
					ShowCustomUi: true,
					Strict:       true,
				},
			},
		},
		{
			SetDataValidation: &sheets.SetDataValidationRequest{
				Range: &sheets.GridRange{
					SheetId:          sheetID,
					StartRowIndex:    int64(consts.TasksStartRow - 1),
					EndRowIndex:      int64(consts.TasksStartRow + rows - 1),
					StartColumnIndex: 4,
					EndColumnIndex:   5,
				},
				Rule: &sheets.DataValidationRule{
					Condition: &sheets.BooleanCondition{
						Type:   "ONE_OF_LIST",
						Values: stateConditionValues,
					},
					ShowCustomUi: true,
					Strict:       true,
				},
			},
		},
	}
}

func (s *SheetsService) createConditionalFormattingRequests(sheetID int64) []*sheets.Request {
	return []*sheets.Request{
		{
			AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
				Rule: &sheets.ConditionalFormatRule{
					Ranges: []*sheets.GridRange{
						{
							SheetId:          sheetID,
							StartRowIndex:    int64(consts.TasksStartRow - 1),
							StartColumnIndex: 2,
							EndColumnIndex:   3,
						},
					},
					GradientRule: &sheets.GradientRule{
						Minpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{
								RgbColor: consts.CostMinColor,
							},
							Type:  "NUMBER",
							Value: "1",
						},
						Midpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{
								RgbColor: consts.CostMidColor,
							},
							Type:  "NUMBER",
							Value: "5",
						},
						Maxpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{
								RgbColor: consts.CostMaxColor,
							},
							Type:  "NUMBER",
							Value: "9",
						},
					},
				},
				Index: 0,
			},
		},
	}
}

func (s *SheetsService) createDimensionRequests(sheetID int64) []*sheets.Request {
	return []*sheets.Request{
		{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: 0,
					EndIndex:   1,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: 80,
				},
				Fields: "pixelSize",
			},
		},
		{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: 1,
					EndIndex:   2,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: 385,
				},
				Fields: "pixelSize",
			},
		},
		{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: 2,
					EndIndex:   3,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: 80,
				},
				Fields: "pixelSize",
			},
		},
		{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: 3,
					EndIndex:   4,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: 150,
				},
				Fields: "pixelSize",
			},
		},
		{
			UpdateDimensionProperties: &sheets.UpdateDimensionPropertiesRequest{
				Range: &sheets.DimensionRange{
					SheetId:    sheetID,
					Dimension:  "COLUMNS",
					StartIndex: 4,
					EndIndex:   5,
				},
				Properties: &sheets.DimensionProperties{
					PixelSize: 100,
				},
				Fields: "pixelSize",
			},
		},
	}
}
