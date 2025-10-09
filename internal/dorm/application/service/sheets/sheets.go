package sheets

import (
	"dorm/internal/common/utils"
	"dorm/internal/dorm/application/model"
	"dorm/internal/dorm/infrastructure"
	"fmt"
)

type SheetsService interface {
	CreateDutySheet(sheetData model.SheetData) error
	UpdateDutySheet(sheetData model.SheetData) error
}

type sheetsService struct {
	sheets  *infrastructure.Sheets
	builder *DutySheetBuilder
	styler  *DutySheetStyler
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

	layout := s.builder.Build(sheetData.Order, sheetData.Tasks, sheetData.Users)

	err = s.sheets.WriteRange(sheetData, layout.TaskTableRange, layout.TaskTableData)
	if err != nil {
		return err
	}
	err = s.sheets.WriteRange(sheetData, layout.UserTableRange, layout.UserTableData)
	if err != nil {
		return fmt.Errorf("failed to write user table: %w", err)
	}

	styleRequests := s.styler.GenerateRequests(sheet.Properties.SheetId, layout, teamColor, sheetData.Users)

	return s.sheets.BatchUpdate(sheetData, styleRequests)
}

func (s *sheetsService) UpdateDutySheet(sheetData model.SheetData) error {
	layout := s.builder.BuildForUpdate(sheetData.Tasks)

	err := s.sheets.ClearRange(sheetData, layout.TaskTableClearRange)
	if err != nil {
		return fmt.Errorf("failed to clear range in sheet '%s': %w", sheetData.Title, err)
	}

	err = s.sheets.WriteRange(sheetData, layout.TaskTableWriteRange, layout.TaskTableData)
	if err != nil {
		return fmt.Errorf("failed to write new data to sheet '%s': %w", sheetData.Title, err)
	}

	return nil
}
