package spreadsheet

import (
	"context"
	"fmt"
	"slices"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

const (
	scope = "https://www.googleapis.com/auth/spreadsheets"
)

type Client interface {
	CreateSheet(spreadsheetID string, title string) (int64, error)
	RecreateSheet(spreadsheetID string, title string) (int64, error)
	HideSheet(spreadsheetID string, sheetID int64) error
	UpdateValues(spreadsheetID string, rangeName string, values [][]interface{}) error
	BatchUpdateValues(spreadsheetID string, data []*sheets.ValueRange) error
	HideSheetsExcept(spreadsheetID string, sheetIDs []int64) error
	BatchUpdate(spreadsheetID string, requests []*sheets.Request) error
}

type SpreadsheetClient struct {
	service *sheets.Service
}

func NewSpreadsheetClient(credentialsJSON string) (*SpreadsheetClient, error) {
	ctx := context.Background()
	credentials, err := google.CredentialsFromJSON(ctx, []byte(credentialsJSON), scope)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w", err)
	}
	service, err := sheets.NewService(ctx, option.WithCredentials(credentials))
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve Sheets client: %w", err)
	}

	return &SpreadsheetClient{service: service}, nil
}

func (c *SpreadsheetClient) CreateSheet(spreadsheetID, title string) (int64, error) {
	request := &sheets.Request{
		AddSheet: &sheets.AddSheetRequest{
			Properties: &sheets.SheetProperties{
				Title: title,
			},
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{request},
	}

	response, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to create sheet: %w", err)
	}

	return response.Replies[0].AddSheet.Properties.SheetId, nil
}

func (c *SpreadsheetClient) RecreateSheet(spreadsheetID string, title string) (int64, error) {
	book, err := c.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	for _, sh := range book.Sheets {
		if sh.Properties != nil && sh.Properties.Title == title {
			_, err = c.service.Spreadsheets.BatchUpdate(
				spreadsheetID,
				&sheets.BatchUpdateSpreadsheetRequest{
					Requests: []*sheets.Request{{
						DeleteSheet: &sheets.DeleteSheetRequest{
							SheetId: sh.Properties.SheetId,
						},
					}},
				},
			).Do()
			if err != nil {
				return 0, fmt.Errorf("failed to delete existing sheet: %w", err)
			}
			break
		}
	}

	return c.CreateSheet(spreadsheetID, title)
}

func (c *SpreadsheetClient) HideSheet(spreadsheetID string, sheetID int64) error {
	request := &sheets.Request{
		UpdateSheetProperties: &sheets.UpdateSheetPropertiesRequest{
			Properties: &sheets.SheetProperties{
				SheetId: sheetID,
				Hidden:  true,
			},
			Fields: "hidden",
		},
	}

	batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: []*sheets.Request{request},
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to hide sheet: %w", err)
	}

	return nil
}

func (c *SpreadsheetClient) UpdateValues(
	spreadsheetID string,
	updateRange string,
	values [][]interface{},
) error {
	_, err := c.service.Spreadsheets.Values.Update(spreadsheetID, updateRange, &sheets.ValueRange{
		Values: values,
	}).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return fmt.Errorf("failed to update values: %w", err)
	}
	return nil
}

func (c *SpreadsheetClient) BatchUpdateValues(spreadsheetID string, data []*sheets.ValueRange) error {
	batchUpdateRequest := &sheets.BatchUpdateValuesRequest{
		ValueInputOption: "USER_ENTERED",
		Data:             data,
	}

	_, err := c.service.Spreadsheets.Values.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
	if err != nil {
		return fmt.Errorf("failed to batch update values: %w", err)
	}

	return nil
}

func (c *SpreadsheetClient) HideSheetsExcept(spreadsheetID string, sheetIDs []int64) error {
	spreadsheet, err := c.service.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		return fmt.Errorf("failed to get spreadsheet: %w", err)
	}

	var requests []*sheets.Request
	for _, sheet := range spreadsheet.Sheets {
		if !slices.Contains(sheetIDs, sheet.Properties.SheetId) {
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

		_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
		if err != nil {
			return fmt.Errorf("failed to hide sheets: %w", err)
		}
	}

	return nil
}

func (c *SpreadsheetClient) BatchUpdate(spreadsheetID string, reqs []*sheets.Request) error {
	batchUpdateReq := &sheets.BatchUpdateSpreadsheetRequest{
		Requests: reqs,
	}

	_, err := c.service.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateReq).Do()
	if err != nil {
		return fmt.Errorf("failed to batch update: %w", err)
	}

	return nil
}
