package types

import (
	"google.golang.org/api/sheets/v4"
)

type Anchor struct {
	Row int64
	Col int64
}

type RenderResult struct {
	Values   [][]interface{}
	Requests []*sheets.Request
}

type Widget interface {
	GetWidth() int64
	Render(sheetID int64, anchor Anchor) RenderResult
}

type ReportDefinition interface {
	GetWidgets() []Widget
	GetTitle() string
	GetSpreadsheetID() string
}
