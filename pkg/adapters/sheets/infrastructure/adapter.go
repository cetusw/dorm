package infrastructure

import (
	"context"
	"log"

	"dorm/pkg/adapters/sheets/app"
	"dorm/pkg/adapters/sheets/domain/reports"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/spreadsheet"
)

type SpreadsheetAdapter struct {
	cleaningUseCase   ports.CleaningUseCase
	userUseCase       ports.UserUseCase
	spreadsheetClient spreadsheet.Client
	eventBus          ports.EventBus
	config            *config.AppConfig
}

func NewSpreadsheetAdapter(
	cleaningUseCase ports.CleaningUseCase,
	userUseCase ports.UserUseCase,
	spreadsheetClient spreadsheet.Client,
	eventBus ports.EventBus,
	config *config.AppConfig,
) (*SpreadsheetAdapter, error) {
	a := &SpreadsheetAdapter{
		cleaningUseCase:   cleaningUseCase,
		userUseCase:       userUseCase,
		spreadsheetClient: spreadsheetClient,
		eventBus:          eventBus,
		config:            config,
	}

	a.eventBus.Subscribe(events.TopicWeekStarted, a.onWeekStarted)

	return a, nil
}

func (a *SpreadsheetAdapter) onWeekStarted(_ context.Context, event interface{}) error {
	e, ok := event.(events.WeekStartedEvent)
	if !ok {
		return nil
	}

	composer := app.NewReportComposer(a.spreadsheetClient)

	for _, duty := range e.Duties {
		sheetID, _ := a.spreadsheetClient.CreateSheet(duty.SpreadsheetID, duty.DutyName)

		report := reports.NewWeeklyReportBlueprint(duty)

		err := composer.Compose(sheetID, report)
		if err != nil {
			log.Printf("Compose error: %v", err)
		}

		_ = a.spreadsheetClient.HideSheetsExcept(duty.SpreadsheetID, []int64{sheetID})
	}
	return nil
}
