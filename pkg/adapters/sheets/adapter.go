package sheets

import (
	"context"
	"fmt"
	"log"

	"dorm/pkg/adapters/sheets/internal/layout"
	"dorm/pkg/core/domain/events"
	"dorm/pkg/core/ports"
	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/gsheets"
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

func (a *Adapter) onWeekStarted(ctx context.Context, event interface{}) error {
	duties, err := a.cleaningUseCase.GetLatestDuties(ctx)
	if err != nil {
		return fmt.Errorf("duties not found: %w", err)
	}

	if len(duties) == 0 {
		log.Println("No duties for the new week, skipping sheet creation.")
		return nil
	}

	for _, concreteDuty := range duties {
		sheetData := layout.BuildSheetData(concreteDuty)

		_, err := a.gsheetsClient.CreateSheet(concreteDuty.SpreadsheetID, concreteDuty.DutyName)
		if err != nil {
			log.Printf("Failed to create sheet for team %s: %v", concreteDuty.TeamName, err)
			continue
		}

		err = a.gsheetsClient.UpdateValues(
			concreteDuty.SpreadsheetID,
			fmt.Sprintf("'%s'!A1", concreteDuty.DutyName),
			sheetData,
		)
		if err != nil {
			log.Printf("Failed to update values for team %s: %v", concreteDuty.TeamName, err)
		}
	}

	return nil
}
