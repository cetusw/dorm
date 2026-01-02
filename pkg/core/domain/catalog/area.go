package catalog

import "context"

type Area struct {
	id    int
	name  string
	floor int
}

func RestoreArea(id int, name string, floor int) *Area {
	return &Area{id: id, name: name, floor: floor}
}

func (a *Area) ID() int      { return a.id }
func (a *Area) Name() string { return a.name }
func (a *Area) Floor() int   { return a.floor }

type AreaRepository interface {
	GetAllAreas(ctx context.Context) ([]*Area, error)
}
