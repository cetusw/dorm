package infrastructure

import (
	"context"
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

	return err
}

func (c *Sheets) UpdateCell(sheetName, cell, value string) error {
	// TODO: дописать логику для обновления ячейки
	return nil
}
