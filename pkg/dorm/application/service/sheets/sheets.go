package sheets

import (
	"fmt"

	"dorm/pkg/common/consts"
	"dorm/pkg/common/utils"
	"dorm/pkg/dorm/application/model"
	"dorm/pkg/dorm/infrastructure"
	infrasheets "dorm/pkg/dorm/infrastructure/sheets"

	"google.golang.org/api/sheets/v4"
)

type SheetsService interface {
	CreateDutySheet(sheetData model.SheetData) error
	UpdateDutySheet(sheetData model.SheetData) error
}

type sheetsService struct {
	sheets         *infrastructure.Sheets
	builder        *DutySheetBuilder
	styler         *DutySheetStyler
	requestFactory *infrasheets.RequestFactory
}

func NewSheetsService(s *infrastructure.Sheets, builder *DutySheetBuilder, styler *DutySheetStyler) SheetsService {
	return &sheetsService{
		sheets:  s,
		builder: builder,
		styler:  styler,
	}
}

func (s *sheetsService) CreateDutySheet(sheetData model.SheetData) error {
	teamColor, err := utils.HexToSheetsColor(sheetData.TeamColor)
	if err != nil {
		return err
	}
	sheet, err := s.sheets.CreateSheet(sheetData.Title, teamColor, sheetData.SpreadsheetID)
	if err != nil {
		return err
	}
	requests, err := s.hideOtherSheetsRequest(sheetData, sheet.Properties.SheetId)
	if err != nil {
		return err
	}

	layout := s.builder.Build(sheetData.Order, sheetData.Tasks, sheetData.Users)

	err = s.sheets.WriteRange(sheetData, consts.TasksTableRange, layout.TaskTableData)
	if err != nil {
		return err
	}
	err = s.sheets.WriteRange(sheetData, consts.UserTableRange, layout.UserTableData)
	if err != nil {
		return fmt.Errorf("failed to write user table: %w", err)
	}

	styleRequests := s.styler.GenerateRequests(sheet.Properties.SheetId, layout, teamColor, sheetData.Users)
	requests = append(requests, styleRequests...)

	return s.sheets.BatchUpdate(sheetData, requests)
}

func (s *sheetsService) UpdateDutySheet(sheetData model.SheetData) error {
	layout := s.builder.BuildForUpdate(sheetData)

	err := s.sheets.ClearRange(sheetData, consts.ClearRange)
	if err != nil {
		return fmt.Errorf("failed to clear range in sheet '%s': %w", sheetData.Title, err)
	}

	err = s.sheets.WriteRange(sheetData, consts.TasksTableRange, layout.TaskTableData)
	if err != nil {
		return fmt.Errorf("failed to write new data to sheet '%s': %w", sheetData.Title, err)
	}
	err = s.sheets.WriteRange(sheetData, consts.UserTableRange, layout.UserTableData)
	if err != nil {
		return fmt.Errorf("failed to write new data to sheet '%s': %w", sheetData.Title, err)
	}

	return nil
}

func (s *sheetsService) hideOtherSheetsRequest(
	sheetData model.SheetData,
	currentSheetID int64,
) ([]*sheets.Request, error) {
	spreadsheet, err := s.sheets.GetSpreadsheet(sheetData.SpreadsheetID)
	if err != nil {
		fmt.Printf("Warning: failed to fetch spreadsheet: %v\n", err)
		return nil, err
	}
	hideRequests := s.requestFactory.BuildHideOtherSheetsRequests(spreadsheet, currentSheetID)
	return hideRequests, nil
}
