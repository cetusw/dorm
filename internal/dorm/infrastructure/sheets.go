package infrastructure

import (
	"context"
	"dorm/internal/dorm/application/model"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheets struct {
	srv *sheets.Service
}

func NewSheets(credentialsFile string) (*Sheets, error) {
	ctx := context.Background()
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, err
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		return nil, err
	}

	client := config.Client(ctx)
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return &Sheets{srv: srv}, nil
}

func (c *Sheets) CreateSheet(title string, color *sheets.Color, spreadsheetID string) (*sheets.Sheet, error) {
	properties := &sheets.SheetProperties{Title: title}
	if color != nil {
		properties.TabColor = color
	}
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{Properties: properties},
	}
	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests:                     []*sheets.Request{req},
		IncludeSpreadsheetInResponse: true,
	}
	resp, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create sheet: %w", err)
	}
	for _, s := range resp.UpdatedSpreadsheet.Sheets {
		if s.Properties.Title == title {
			return s, nil
		}
	}

	return nil, fmt.Errorf("cannot find sheet '%s'", title)
}

func (c *Sheets) WriteRange(sheetData model.SheetData, startCell string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	rangeStr := sheetData.Title + "!" + startCell
	_, err := c.srv.Spreadsheets.Values.Update(sheetData.SpreadsheetID, rangeStr, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (c *Sheets) ClearRange(sheetData model.SheetData, clearRange string) error {
	fullRange := fmt.Sprintf("%s!%s", sheetData.Title, clearRange)

	clearRequest := &sheets.ClearValuesRequest{}

	_, err := c.srv.Spreadsheets.Values.Clear(sheetData.SpreadsheetID, fullRange, clearRequest).Do()
	if err != nil {
		return fmt.Errorf("unable to clear range %s: %w", fullRange, err)
	}

	return nil
}

func (c *Sheets) ReadSheet(spreadsheetID string, readRange string) ([][]interface{}, error) {
	response, err := c.srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve data from sheet: %w", err)
	}
	return response.Values, nil
}

func (c *Sheets) BatchUpdate(sheetData model.SheetData, requests []*sheets.Request) error {
	if len(requests) == 0 {
		return nil
	}
	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}
	_, err := c.srv.Spreadsheets.BatchUpdate(sheetData.SpreadsheetID, batchUpdateReq).Do()
	return err
}

func (c *Sheets) GetSpreadsheet(spreadsheetID string) (*sheets.Spreadsheet, error) {
	resp, err := c.srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get spreadsheet: %w", err)
	}
	return resp, nil
}
