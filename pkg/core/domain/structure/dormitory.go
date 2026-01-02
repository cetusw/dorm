package structure

import (
	"context"

	"github.com/google/uuid"
)

type Dormitory struct {
	id       int64
	name     string
	leaderID *uuid.UUID
	city     string
}

func RestoreDormitory(id int64, name string, leaderID *uuid.UUID, city string) *Dormitory {
	return &Dormitory{
		id:       id,
		name:     name,
		leaderID: leaderID,
		city:     city,
	}
}

func (d *Dormitory) ID() int64            { return d.id }
func (d *Dormitory) Name() string         { return d.name }
func (d *Dormitory) LeaderID() *uuid.UUID { return d.leaderID }

type DormitoryRepository interface {
	FindAll(ctx context.Context) ([]*Dormitory, error)
	FindByID(ctx context.Context, id int64) (*Dormitory, error)
}
