package sheets

import (
	"dorm/pkg/common/consts"
	"dorm/pkg/common/utils"
	"dorm/pkg/dorm/application/model"
	"fmt"

	"google.golang.org/api/sheets/v4"
)

type DutySheetStyler struct {
	formatter *SheetFormatter
}

func NewDutySheetStyler() *DutySheetStyler {

	return &DutySheetStyler{}
}

func (s *DutySheetStyler) GenerateRequests(sheetId int64, layout *SheetLayout, teamColor *sheets.Color, users []model.User) []*sheets.Request {
	s.formatter = NewSheetFormatter(sheetId)
	var allRequests []*sheets.Request

	taskRequests := s.createTaskTableFormattingRequests(layout, teamColor, users)
	allRequests = append(allRequests, taskRequests...)

	userRequests := s.createUserTableFormattingRequests(layout.UserCount, teamColor)
	allRequests = append(allRequests, userRequests...)

	return allRequests
}

func (s *DutySheetStyler) createUserTableFormattingRequests(userRowCount int, teamColor *sheets.Color) []*sheets.Request {
	var requests []*sheets.Request

	requests = append(requests, s.getUserTableHeaderRequest()...)
	requests = append(requests, s.getUserTableRowsRequest(int64(userRowCount+1), teamColor)...)
	requests = append(requests, s.getUserTableColumnsWidthRequest([]int64{190, 90, 90}, 6)...)

	return requests
}

func (s *DutySheetStyler) getUserTableHeaderRequest() []*sheets.Request {
	return []*sheets.Request{
		s.formatter.SetBackgroundColor(1, 2, 6, 9, consts.SubHeaderBgColor),
		s.formatter.SetTextAlignment(1, 2, 6, 9, "CENTER", "MIDDLE"),
		s.formatter.SetTextFormat(1, 2, 6, 9, true, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
	}
}

func (s *DutySheetStyler) getUserTableRowsRequest(rowCount int64, teamColor *sheets.Color) []*sheets.Request {
	if rowCount <= 2 {
		return []*sheets.Request{}
	}

	return []*sheets.Request{
		s.formatter.SetOuterBorders(1, 2, 6, 9),
		s.formatter.SetOuterBorders(2, rowCount, 6, 9),
		s.formatter.SetOuterBorders(1, rowCount, 7, 9),
		s.formatter.SetOuterBorders(rowCount, rowCount+1, 6, 9),
		s.formatter.MergeCells(rowCount, rowCount+1, 7, 9, "MERGE_ALL"),
		s.formatter.SetTextAlignment(rowCount, rowCount+1, 7, 9, "LEFT", "MIDDLE"),
		s.formatter.SetTextFormat(rowCount, rowCount+1, 6, 9, true, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
		s.formatter.SetBackgroundColor(2, rowCount, 6, 7, teamColor),
		s.formatter.SetTextAlignment(2, rowCount, 6, 9, "LEFT", "MIDDLE"),
		s.formatter.SetTextFormat(2, rowCount, 6, 9, false, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
	}
}

func (s *DutySheetStyler) getUserTableColumnsWidthRequest(columnWidths []int64, startColumn int64) []*sheets.Request {
	var requests []*sheets.Request
	for i, width := range columnWidths {
		requests = append(requests, s.formatter.SetColumnWidth(startColumn+int64(i), width))
	}
	return requests
}

func (s *DutySheetStyler) createTaskTableFormattingRequests(layout *SheetLayout, teamColor *sheets.Color, users []model.User) []*sheets.Request {
	var requests []*sheets.Request

	requests = append(requests, s.getTaskTableHeaderRequest(teamColor)...)
	requests = append(requests, s.getTaskTableSubHeaderRequest()...)
	requests = append(requests, s.getTaskTableRowsRequest(layout.ZoneMergeRange, layout.TaskCount)...)
	requests = append(requests, s.getValidationRequest(layout.TaskCount, users)...)
	requests = append(requests, s.getConditionalFormattingRequest()...)
	requests = append(requests, s.getTaskTableColumnsWidthRequest()...)

	return requests
}

func (s *DutySheetStyler) getTaskTableHeaderRequest(teamColor *sheets.Color) []*sheets.Request {
	return []*sheets.Request{
		s.formatter.MergeCells(0, 1, 0, 5, "MERGE_ALL"),
		s.formatter.SetOuterBorders(consts.TaskTableStartRow, 1, 0, 5),
		s.formatter.SetRowHeight(consts.TaskTableStartRow, consts.TaskTableHeaderHeight),
		s.formatter.SetBackgroundColor(0, 1, 0, 5, teamColor),
		s.formatter.SetTextAlignment(0, 1, 0, 5, "CENTER", "MIDDLE"),
		s.formatter.SetTextFormat(
			0,
			1,
			0,
			5,
			true,
			consts.TaskTableHeaderFontSize,
			consts.DefaultFontFamily,
			utils.DarkenColorRGB(teamColor, 0.5),
		),
	}
}

func (s *DutySheetStyler) getTaskTableSubHeaderRequest() []*sheets.Request {
	return []*sheets.Request{
		s.formatter.SetOuterBorders(1, 2, 0, 5),
		s.formatter.SetBackgroundColor(1, 2, 0, 5, consts.SubHeaderBgColor),
		s.formatter.SetTextAlignment(1, 2, 0, 5, "CENTER", "MIDDLE"),
		s.formatter.SetTextFormat(1, 2, 0, 5, true, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
	}
}

func (s *DutySheetStyler) getTaskTableRowsRequest(zoneMergeRanges map[string][2]int, rows int) []*sheets.Request {
	startRowTasks := int64(consts.TasksStartRow - 1)
	endRowTasks := startRowTasks + int64(rows)
	var requests []*sheets.Request
	requests = s.getAreaMergeRequest(zoneMergeRanges)
	requests = append(requests, s.getAreaOuterBordersRequest(zoneMergeRanges)...)

	requests = append(requests, []*sheets.Request{
		s.formatter.SetOuterBorders(1, startRowTasks+int64(rows), 1, 2),
		s.formatter.SetOuterBorders(1, startRowTasks+int64(rows), 2, 3),
		s.formatter.SetTextAlignment(startRowTasks, endRowTasks, 2, 3, "CENTER", "MIDDLE"),
		s.formatter.SetBackgroundColor(startRowTasks, endRowTasks, 0, 5, consts.TasksBackgroundColor),
		s.formatter.SetTextAlignment(startRowTasks, endRowTasks, 0, 1, "CENTER", "MIDDLE"),
		s.formatter.SetWrapStrategy(startRowTasks, endRowTasks, 0, 1, "WRAP"),
		s.formatter.SetTextFormat(startRowTasks, endRowTasks, 0, 1, true, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
		s.formatter.SetTextFormat(startRowTasks, endRowTasks, 1, 5, false, consts.DefaultFontSize, consts.DefaultFontFamily, nil),
		s.formatter.SetTextRotation(startRowTasks, endRowTasks, 0, 1, 90),
	}...)

	return requests
}

func (s *DutySheetStyler) getAreaOuterBordersRequest(zoneMergeRanges map[string][2]int) []*sheets.Request {
	var requests []*sheets.Request
	for _, ranges := range zoneMergeRanges {
		if ranges[0] >= ranges[1] {
			continue
		}
		startRow := int64(ranges[0] - 1)
		endRow := int64(ranges[1])
		requests = append(requests, s.formatter.SetOuterBorders(startRow, endRow, 0, 5))
	}
	return requests
}

func (s *DutySheetStyler) getAreaMergeRequest(zoneMergeRanges map[string][2]int) []*sheets.Request {
	var requests []*sheets.Request
	for _, ranges := range zoneMergeRanges {
		if ranges[0] >= ranges[1] {
			continue
		}
		startRow := int64(ranges[0] - 1)
		endRow := int64(ranges[1])
		requests = append(requests, s.formatter.MergeCells(startRow, endRow, 0, 1, "MERGE_COLUMNS"))
	}
	return requests
}

func (s *DutySheetStyler) getValidationRequest(rows int, users []model.User) []*sheets.Request {
	var userValues []string
	userValues = append(userValues, consts.DefaultAssignee)
	for _, user := range users {
		userName := fmt.Sprintf("%s %s.", user.FirstName, string([]rune(user.LastName)[0]))
		userValues = append(userValues, userName)
	}

	stateValues := []string{consts.StateNotDone, consts.StateDone, consts.StateVerified}

	startRow := int64(consts.TasksStartRow - 1)
	endRow := int64(consts.TasksStartRow + rows - 1)

	return []*sheets.Request{
		s.formatter.SetDataValidation(startRow, endRow, 3, 4, userValues),
		s.formatter.SetDataValidation(startRow, endRow, 4, 5, stateValues),
	}
}

func (s *DutySheetStyler) getConditionalFormattingRequest() []*sheets.Request {
	return []*sheets.Request{
		{
			AddConditionalFormatRule: &sheets.AddConditionalFormatRuleRequest{
				Rule: &sheets.ConditionalFormatRule{
					Ranges: []*sheets.GridRange{
						{
							SheetId:          s.formatter.GetID(),
							StartRowIndex:    int64(consts.TasksStartRow - 1),
							StartColumnIndex: 2,
							EndColumnIndex:   3,
						},
					},
					GradientRule: &sheets.GradientRule{
						Minpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{RgbColor: consts.CostMinColor},
							Type:       "NUMBER", Value: "1",
						},
						Midpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{RgbColor: consts.CostMidColor},
							Type:       "NUMBER", Value: "5",
						},
						Maxpoint: &sheets.InterpolationPoint{
							ColorStyle: &sheets.ColorStyle{RgbColor: consts.CostMaxColor},
							Type:       "NUMBER", Value: "9",
						},
					},
				},
				Index: 0,
			},
		},
	}
}

func (s *DutySheetStyler) getTaskTableColumnsWidthRequest() []*sheets.Request {
	return []*sheets.Request{
		s.formatter.SetColumnWidth(0, 80),
		s.formatter.SetColumnWidth(1, 385),
		s.formatter.SetColumnWidth(2, 100),
		s.formatter.SetColumnWidth(3, 150),
		s.formatter.SetColumnWidth(4, 120),
	}
}
