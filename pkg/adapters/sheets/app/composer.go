package app

import (
	"fmt"

	"google.golang.org/api/sheets/v4"

	"dorm/pkg/adapters/sheets/domain"
	"dorm/pkg/adapters/sheets/types"
	"dorm/pkg/infrastructure/spreadsheet"
)

type ReportComposer struct {
	client spreadsheet.Client
	engine *domain.LayoutEngine
}

func NewReportComposer(client spreadsheet.Client) *ReportComposer {
	return &ReportComposer{
		client: client,
		engine: &domain.LayoutEngine{Margin: 1},
	}
}

func (c *ReportComposer) Compose(sheetID int64, report types.ReportDefinition) error {
	widgets := report.GetWidgets()
	anchors := c.engine.CalculateAnchors(widgets)

	var allRequests []*sheets.Request

	for _, w := range widgets {
		anchor := anchors[w]
		render := w.Render(sheetID, anchor)

		rangeName := fmt.Sprintf("%s!%s%d", report.GetTitle(), c.colIndexToLetter(anchor.Col), anchor.Row+1)
		err := c.client.UpdateValues(report.GetSpreadsheetID(), rangeName, render.Values)
		if err != nil {
			return err
		}

		allRequests = append(allRequests, render.Requests...)
	}

	return c.client.BatchUpdate(report.GetSpreadsheetID(), allRequests)
}

func (c *ReportComposer) colIndexToLetter(col int64) string {
	return string(rune('A' + col))
}
