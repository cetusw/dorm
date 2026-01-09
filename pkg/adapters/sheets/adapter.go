package sheets

import (
	"context"
	"dorm/pkg/adapters/sheets/internal/layout"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/gsheets"

	"google.golang.org/api/sheets/v4"
)

type Adapter struct {
	cleaningUseCase ports.CleaningUseCase
	userUseCase     ports.UserUseCase
	gsheetsClient   gsheets.SpreadsheetClient
	eventBus        ports.EventBus
	config          *config.AppConfig
}

func NewAdapter(
	cleaningUseCase ports.CleaningUseCase,
	userUseCase ports.UserUseCase,
	gsheetsClient gsheets.SpreadsheetClient,
	eventBus ports.EventBus,
	config *config.AppConfig,
) (*Adapter, error) {
	a := &Adapter{
		cleaningUseCase: cleaningUseCase,
		userUseCase:     userUseCase,
		gsheetsClient:   gsheetsClient,
		eventBus:        eventBus,
		config:          config,
	}

	a.eventBus.Subscribe(events.TopicWeekStarted, a.onWeekStarted)

	return a, nil
}

func (a *Adapter) onWeekStarted(ctx context.Context, _ interface{}) error {
	duties, _ := a.cleaningUseCase.GetLatestDuties(ctx)

	for _, d := range duties {
		sheetID, _ := a.gsheetsClient.CreateSheet(d.SpreadsheetID, d.DutyName)

		mainData := layout.BuildSheetData(d)
		statsData := layout.PrepareStatsTable(d)

		err := a.gsheetsClient.UpdateValues(d.SpreadsheetID, d.DutyName+"!A1", mainData)
		if err != nil {
			return err
		}
		err = a.gsheetsClient.UpdateValues(d.SpreadsheetID, d.DutyName+"!G3", statsData)
		if err != nil {
			return err
		}

		requests := []*sheets.Request{
			ApplyHeaderStyle(sheetID, d.TeamColor),
			AddCostGradient(sheetID, len(d.Tasks)),
			AddExecutorValidation(sheetID, d.UsersStats, len(d.Tasks)),
			AddStatusValidation(sheetID, len(d.Tasks)),
		}
		requests = append(requests, MergeAreaCells(sheetID, d.Tasks)...)

		err = a.gsheetsClient.BatchUpdate(d.SpreadsheetID, requests)
		if err != nil {
			return err
		}
		err = a.gsheetsClient.HideSheetsExcept(d.SpreadsheetID, []int64{sheetID})
		if err != nil {
			return err
		}
	}
	return nil
}
