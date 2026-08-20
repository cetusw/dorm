package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type PenaltyScope struct {
	DormitoryIDs []int64
}

type PenaltyQueryService interface {
	ListResidentsWithPenaltyBalance(
		ctx context.Context,
		scope PenaltyScope,
	) ([]dto.PenaltyResidentSummary, error)

	SearchEligibleResidents(
		ctx context.Context,
		scope PenaltyScope,
		search string,
	) ([]dto.PenaltyResidentOption, error)

	ListPenaltyEntriesByUser(
		ctx context.Context,
		userID uuid.UUID,
	) ([]dto.PenaltyEntryItem, error)
}
