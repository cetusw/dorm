package structure

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

type Dormitory struct {
	id          int64
	name        string
	leaderID    *uuid.UUID
	city        string
	streetType  string
	streetName  string
	houseNumber string
}

func NewDormitory(name, city, streetType, streetName, houseNumber string) *Dormitory {
	return &Dormitory{
		name:        name,
		city:        city,
		streetType:  streetType,
		streetName:  streetName,
		houseNumber: houseNumber,
	}
}

func RestoreDormitory(id int64, name string, leaderID *uuid.UUID, city, streetType, streetName, houseNumber string) *Dormitory {
	return &Dormitory{
		id:          id,
		name:        name,
		leaderID:    leaderID,
		city:        city,
		streetType:  streetType,
		streetName:  streetName,
		houseNumber: houseNumber,
	}
}

func (d *Dormitory) ID() int64            { return d.id }
func (d *Dormitory) Name() string         { return d.name }
func (d *Dormitory) LeaderID() *uuid.UUID { return d.leaderID }
func (d *Dormitory) City() string         { return d.city }
func (d *Dormitory) StreetType() string   { return d.streetType }
func (d *Dormitory) StreetName() string   { return d.streetName }
func (d *Dormitory) HouseNumber() string  { return d.houseNumber }
func (d *Dormitory) Address() string {
	street := strings.TrimSpace(d.streetType + " " + d.streetName)
	parts := make([]string, 0, 3)
	for _, part := range []string{d.city, street, d.houseNumber} {
		if strings.TrimSpace(part) != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ", ")
}

type DormitoryRepository interface {
	FindAll(ctx context.Context) ([]*Dormitory, error)
	FindByID(ctx context.Context, id int64) (*Dormitory, error)
	ExistsByLeaderID(ctx context.Context, leaderID uuid.UUID) (bool, error)
	Save(ctx context.Context, dormitory *Dormitory) error
	Delete(ctx context.Context, id int64) error
}
