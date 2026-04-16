package infrastructure

import (
	"context"
	"fmt"
	"log"

	"dorm/pkg/adapters/sheets/app"
	"dorm/pkg/adapters/sheets/domain/reports"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/ports"
	"dorm/pkg/core/ports/dto"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/spreadsheet"

	"github.com/google/uuid"
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
		sheetID, err := a.spreadsheetClient.RecreateSheet(duty.SpreadsheetID, duty.DutyName)
		if err != nil {
			return fmt.Errorf("failed to create sheet: %w", err)
		}

		report := reports.NewWeeklyReportBlueprint(duty)

		err = composer.Compose(sheetID, report)
		if err != nil {
			log.Printf("Compose error: %v", err)
		}

		_ = a.spreadsheetClient.HideSheetsExcept(duty.SpreadsheetID, []int64{sheetID})
	}
	return nil
}

func (a *SpreadsheetAdapter) RegenerateCurrentDutySheet(ctx context.Context, teamID uuid.UUID) error {
	duties, err := a.cleaningUseCase.GetLatestDuties(ctx)
	if err != nil {
		return err
	}
	var dutyToRegen *dto.DutyViewModel
	for i := range duties {
		if duties[i].TeamID == teamID {
			dutyToRegen = &duties[i]
			break
		}
	}
	if dutyToRegen == nil {
		return fmt.Errorf("current duty for team not found")
	}

	composer := app.NewReportComposer(a.spreadsheetClient)
	sheetID, err := a.spreadsheetClient.RecreateSheet(dutyToRegen.SpreadsheetID, dutyToRegen.DutyName)
	if err != nil {
		return fmt.Errorf("failed to recreate sheet: %w", err)
	}
	report := reports.NewWeeklyReportBlueprint(*dutyToRegen)
	if err := composer.Compose(sheetID, report); err != nil {
		return fmt.Errorf("failed to compose report: %w", err)
	}
	_ = a.spreadsheetClient.HideSheetsExcept(dutyToRegen.SpreadsheetID, []int64{sheetID})
	return nil
}
