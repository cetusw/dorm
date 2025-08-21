package infrastructure

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheets struct {
	srv           *sheets.Service
	spreadsheetID string
}

func NewSheets(credentialsFile string, spreadsheetID string) (*Sheets, error) {
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

	return &Sheets{srv: srv, spreadsheetID: spreadsheetID}, nil
}

func (c *Sheets) CreateSheet(title string, color *sheets.Color) error {
	properties := &sheets.SheetProperties{Title: title}
	if color != nil {
		properties.TabColor = color
	}

	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{Properties: properties},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err := c.srv.Spreadsheets.BatchUpdate(c.spreadsheetID, batchUpdateReq).Do()
	return err
}

func (c *Sheets) WriteRange(sheetTitle, startCell string, data [][]interface{}) error {
	valueRange := &sheets.ValueRange{
		Values: data,
	}

	rangeStr := sheetTitle + "!" + startCell
	_, err := c.srv.Spreadsheets.Values.Update(c.spreadsheetID, rangeStr, valueRange).
		ValueInputOption("USER_ENTERED").
		Do()

	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (c *Sheets) CreateSheetAndGetID(title string, color *sheets.Color) (*sheets.Sheet, error) {
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
	resp, err := c.srv.Spreadsheets.BatchUpdate(c.spreadsheetID, batchUpdateReq).Do()
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

func (c *Sheets) BatchUpdate(requests []*sheets.Request) error {
	if len(requests) == 0 {
		return nil
	}
	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: requests,
	}
	_, err := c.srv.Spreadsheets.BatchUpdate(c.spreadsheetID, batchUpdateReq).Do()
	return err
}
