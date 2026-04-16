package ports

import (
	"context"

	"github.com/google/uuid"
)

type SheetsUseCase interface {
	RegenerateCurrentDutySheet(ctx context.Context, teamID uuid.UUID) error
}

