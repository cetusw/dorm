package structure

import (
	"context"

	"github.com/google/uuid"
)

type Group struct {
	id            uuid.UUID
	name          string
	spreadsheetID string
	dormitoryID   int64
}

func RestoreGroup(id uuid.UUID, name, spreadsheetID string, dormitoryID int64) *Group {
	return &Group{
		id:            id,
		name:          name,
		spreadsheetID: spreadsheetID,
		dormitoryID:   dormitoryID,
	}
}

func (g *Group) ID() uuid.UUID         { return g.id }
func (g *Group) Name() string          { return g.name }
func (g *Group) SpreadsheetID() string { return g.spreadsheetID }
func (g *Group) DormitoryID() int64    { return g.dormitoryID }

type GroupRepository interface {
	FindAll(ctx context.Context) ([]*Group, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Group, error)
}
