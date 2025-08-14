package infrastructure

import (
	"context"
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

func (c *Sheets) CreateNewSheet(spreadsheetID, title string) error {
	return nil
}

func (c *Sheets) UpdateCell(spreadsheetID, sheetName, cellRange string, value string) error {
	return nil
}
