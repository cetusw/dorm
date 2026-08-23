package ports

import (
	"context"
	"dorm/pkg/core/ports/dto"
	"github.com/google/uuid"
)

type DutyParticipantsUseCase interface {
	GetParticipants(ctx context.Context, actorID, dutyID uuid.UUID) ([]dto.DutyParticipantResponse, error)
	GetCandidates(ctx context.Context, actorID, dutyID uuid.UUID) ([]dto.DutyParticipantCandidateResponse, error)
	AddParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID) error
	ExcludeParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID, replacementLeaderID *uuid.UUID) error
	RestoreParticipant(ctx context.Context, actorID, dutyID, participantID uuid.UUID) error
	ChangeLeader(ctx context.Context, actorID, dutyID, leaderID uuid.UUID) error
}
