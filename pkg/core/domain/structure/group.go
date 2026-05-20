package structure

import (
	"context"

	"github.com/google/uuid"
)

type Group struct {
	id            uuid.UUID
	leaderID      *uuid.UUID
	name          string
	spreadsheetID string
	dormitoryID   int64
	nextDutyTeam  *int
}

func NewGroup(name string, leaderID *uuid.UUID, spreadsheetID string, dormitoryID int64) *Group {
	return &Group{
		id:            uuid.New(),
		leaderID:      leaderID,
		name:          name,
		spreadsheetID: spreadsheetID,
		dormitoryID:   dormitoryID,
	}
}

func RestoreGroup(id uuid.UUID, leaderID *uuid.UUID, name, spreadsheetID string, dormitoryID int64, nextDutyTeam *int) *Group {
	return &Group{
		id:            id,
		leaderID:      leaderID,
		name:          name,
		spreadsheetID: spreadsheetID,
		dormitoryID:   dormitoryID,
		nextDutyTeam:  nextDutyTeam,
	}
}

func (g *Group) SetNextDutyTeam(order *int) { g.nextDutyTeam = order }

func (g *Group) ID() uuid.UUID         { return g.id }
func (g *Group) LeaderID() *uuid.UUID  { return g.leaderID }
func (g *Group) Name() string          { return g.name }
func (g *Group) SpreadsheetID() string { return g.spreadsheetID }
func (g *Group) DormitoryID() int64    { return g.dormitoryID }
func (g *Group) NextDutyTeam() *int    { return g.nextDutyTeam }

type GroupRepository interface {
	FindAll(ctx context.Context) ([]*Group, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Group, error)
	FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*Group, error)
	Save(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id uuid.UUID) error
}
