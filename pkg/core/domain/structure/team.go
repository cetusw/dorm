package structure

import (
	"context"

	"github.com/google/uuid"
)

type Team struct {
	id       uuid.UUID
	groupID  uuid.UUID
	leaderID *uuid.UUID
	color    string
	order    int
}

func RestoreTeam(id, groupID uuid.UUID, leaderID *uuid.UUID, color string, order int) *Team {
	return &Team{
		id:       id,
		groupID:  groupID,
		leaderID: leaderID,
		color:    color,
		order:    order,
	}
}

func (t *Team) ID() uuid.UUID        { return t.id }
func (t *Team) GroupID() uuid.UUID   { return t.groupID }
func (t *Team) LeaderID() *uuid.UUID { return t.leaderID }
func (t *Team) Color() string        { return t.color }
func (t *Team) Order() int           { return t.order }

type TeamRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Team, error)
	FindByGroupID(ctx context.Context, groupID uuid.UUID) ([]*Team, error)
}
