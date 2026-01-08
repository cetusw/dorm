package gsheets

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type SpreadsheetClient interface {
	CreateSheet(spreadsheetID, title string) (int64, error)
	HideSheet(spreadsheetID string, sheetID int64) error
	UpdateValues(spreadsheetID, range_ string, values [][]interface{}) error
	BatchUpdateValues(spreadsheetID string, data []*sheets.ValueRange) error
	HideSheetsExcept(spreadsheetID string, sheetIDs []int64) error
}

type Client struct {
	srv *sheets.Service
}

func NewClient(credentialsJSON string) (*Client, error) {
	ctx := context.Background()
	b, err := google.CredentialsFromJSON(ctx, []byte(credentialsJSON), "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}
	srv, err := sheets.NewService(ctx, option.WithCredentials(b))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	return &Client{srv: srv}, nil
}

func (c *Client) CreateSheet(spreadsheetID, title string) (int64, error) {
	req := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: title,
			},
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	res, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to create sheet: %w", err)
	}

	return res.Replies[0].AddSheet.Properties.SheetId, nil
}

func (c *Client) HideSheet(spreadsheetID string, sheetID int64) error {
	req := &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				Hidden:  true,
			},
			Fields: "hidden",
		},
	}

	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{req},
	}

	_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
	if err != nil {
		return fmt.Errorf("failed to hide sheet: %w", err)
	}

	return nil
}

func (c *Client) UpdateValues(spreadsheetID, range_ string, values [][]interface{}) error {
	_, err := c.srv.Spreadsheets.Values.Update(spreadsheetID, range_, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to update values: %w", err)
	}
	return nil
}

func (c *Client) BatchUpdateValues(spreadsheetID string, data []*sheets.ValueRange) error {
	batchUpdateReq := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}

	_, err := c.srv.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
	if err != nil {
		return fmt.Errorf("failed to batch update values: %w", err)
	}

	return nil
}

func (c *Client) HideSheetsExcept(spreadsheetID string, sheetIDs []int64) error {
	spreadsheet, err := c.srv.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var requests []*sheets.Request
	for _, sheet := range spreadsheet.Sheets {
		shouldHide := true
		for _, id := range sheetIDs {
			if sheet.Properties.SheetId == id {
				shouldHide = false
				break
			}
		}

		if shouldHide {
			requests = append(requests, &sheets.Request{
				UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
					Properties: &sheets.SheetProperties{
						SheetId: sheet.Properties.SheetId,
						Hidden:  true,
					},
					Fields: "hidden",
				},
			})
		}
	}

	if len(requests) > 0 {
		batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: requests,
		}

		_, err := c.srv.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
		if err != nil {
			return fmt.Errorf("failed to hide sheets: %w", err)
		}
	}

	return nil
}
