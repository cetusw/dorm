package sheets

import (
	gsheets "google.golang.org/api/sheets/v4"
)

type RequestFactory struct{}

func NewRequestFactory() *RequestFactory {
	return &RequestFactory{}
}

func (f *RequestFactory) BuildHideOtherSheetsRequests(
	spreadsheet *gsheets.Spreadsheet,
	keepSheetID int64,
) []*gsheets.Request {
	var requests []*gsheets.Request

	if spreadsheet == nil {
		return requests
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.SheetId == keepSheetID || sheet.Properties.Hidden {
			continue
		}

		req := &gsheets.Request{
			UpdateSheetProperties: &gsheets.UpdateSheetPropertiesRequest{
				Properties: &gsheets.SheetProperties{
					SheetId: sheet.Properties.SheetId,
					Hidden:  true,
				},
				Fields: "hidden",
			},
		}
		requests = append(requests, req)
	}

	return requests
}
