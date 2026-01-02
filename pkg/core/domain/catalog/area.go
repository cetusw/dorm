package catalog

import (
	"context"

	"github.com/google/uuid"
)

type Area struct {
	id      int
	name    string
	floor   int
	groupID *uuid.UUID
}

func RestoreArea(id int, name string, floor int, groupID *uuid.UUID) *Area {
	return &Area{id: id, name: name, floor: floor, groupID: groupID}
}

func (a *Area) AssignGroup(groupID *uuid.UUID) { a.groupID = groupID }

func (a *Area) ID() int             { return a.id }
func (a *Area) Name() string        { return a.name }
func (a *Area) Floor() int          { return a.floor }
func (a *Area) GroupID() *uuid.UUID { return a.groupID }

type AreaRepository interface {
	GetAllAreas(ctx context.Context) ([]*Area, error)
}
