package structure

import (
	"context"

	"github.com/google/uuid"
)

type Group struct {
	id          uuid.UUID
	leaderID    *uuid.UUID
	name        string
	dormitoryID int64
}

func NewGroup(name string, leaderID *uuid.UUID, dormitoryID int64) *Group {
	return &Group{
		id:          uuid.New(),
		leaderID:    leaderID,
		name:        name,
		dormitoryID: dormitoryID,
	}
}

func RestoreGroup(id uuid.UUID, leaderID *uuid.UUID, name string, dormitoryID int64) *Group {
	return &Group{
		id:          id,
		leaderID:    leaderID,
		name:        name,
		dormitoryID: dormitoryID,
	}
}

func (g *Group) ID() uuid.UUID        { return g.id }
func (g *Group) LeaderID() *uuid.UUID { return g.leaderID }
func (g *Group) Name() string         { return g.name }
func (g *Group) DormitoryID() int64   { return g.dormitoryID }

type GroupRepository interface {
	FindAll(ctx context.Context) ([]*Group, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Group, error)
	FindByDormitoryID(ctx context.Context, dormitoryID int64) ([]*Group, error)
	Save(ctx context.Context, group *Group) error
	Delete(ctx context.Context, id uuid.UUID) error
}
