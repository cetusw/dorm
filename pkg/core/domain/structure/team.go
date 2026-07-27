package structure

import (
	"context"

	"github.com/google/uuid"
)

type Team struct {
	id               uuid.UUID
	name             string
	groupID          uuid.UUID
	leaderID         *uuid.UUID
	color            string
	rotationPosition int
}

func NewTeam(name string, groupID uuid.UUID, color string, rotationPosition int) *Team {
	return &Team{
		id:               uuid.New(),
		name:             name,
		groupID:          groupID,
		color:            color,
		rotationPosition: normalizeRotationPosition(rotationPosition),
	}
}

func RestoreTeam(
	id uuid.UUID,
	name string,
	groupID uuid.UUID,
	leaderID *uuid.UUID,
	color string,
	rotationPosition int,
) *Team {
	return &Team{
		id:               id,
		name:             name,
		groupID:          groupID,
		leaderID:         leaderID,
		color:            color,
		rotationPosition: normalizeRotationPosition(rotationPosition),
	}
}

func normalizeRotationPosition(rotationPosition int) int {
	if rotationPosition <= 0 {
		return 1
	}
	return rotationPosition
}

func (t *Team) ID() uuid.UUID         { return t.id }
func (t *Team) Name() string          { return t.name }
func (t *Team) GroupID() uuid.UUID    { return t.groupID }
func (t *Team) LeaderID() *uuid.UUID  { return t.leaderID }
func (t *Team) Color() string         { return t.color }
func (t *Team) RotationPosition() int { return t.rotationPosition }

type TeamRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Team, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Team, error)
	UpdateRotationPositions(ctx context.Context, groupID uuid.UUID, orderedTeamIDs []uuid.UUID) error
	Save(ctx context.Context, team *Team) error
	Delete(ctx context.Context, id uuid.UUID) error
}
