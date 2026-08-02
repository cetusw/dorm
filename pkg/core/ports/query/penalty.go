package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type PenaltyScope struct {
	DormitoryIDs []int64
	GroupIDs     []uuid.UUID
}

type PenaltyQueryService interface {
	ListResidentsWithActivePenalties(
		ctx context.Context,
		scope PenaltyScope,
	) ([]dto.PenaltyResidentSummary, error)

	SearchEligibleResidents(
		ctx context.Context,
		scope PenaltyScope,
		search string,
	) ([]dto.PenaltyResidentOption, error)

	ListActivePenaltiesByUser(
		ctx context.Context,
		userID uuid.UUID,
	) ([]dto.PenaltyItem, error)
}
