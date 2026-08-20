package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"

	"github.com/google/uuid"
)

type PenaltyUseCase interface {
	ListResidents(
		ctx context.Context,
		currentUserID uuid.UUID,
	) (dto.PenaltyResidentsResponse, error)

	SearchResidents(
		ctx context.Context,
		currentUserID uuid.UUID,
		search string,
	) (dto.PenaltyResidentOptionsResponse, error)

	GetResidentPenalties(
		ctx context.Context,
		currentUserID uuid.UUID,
		residentID uuid.UUID,
	) (*dto.PenaltyResidentDetailsResponse, error)
	GetCurrentUserPenalties(
		ctx context.Context,
		currentUserID uuid.UUID,
	) (*dto.CurrentUserPenaltiesResponse, error)

	CreatePenalty(
		ctx context.Context,
		currentUserID uuid.UUID,
		request dto.CreatePenaltyRequest,
	) (*dto.PenaltyEntryItem, error)

	ResolvePenalty(
		ctx context.Context,
		currentUserID uuid.UUID,
		request dto.ResolvePenaltyRequest,
	) error

	DeletePenaltyEntry(
		ctx context.Context,
		currentUserID uuid.UUID,
		entryID uuid.UUID,
	) error

	UpdatePenaltyEntry(
		ctx context.Context,
		currentUserID uuid.UUID,
		entryID uuid.UUID,
		request dto.UpdatePenaltyEntryRequest,
	) (*dto.PenaltyEntryItem, error)
}
