package penalty

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	minWeight         = 0.1
	maxWeight         = 999999999.9
	maxReasonRuneSize = 256
)

type Penalty struct {
	id         uuid.UUID
	userID     uuid.UUID
	weight     float64
	reason     string
	issuedOn   time.Time
	createdAt  time.Time
	resolvedAt *time.Time
}

func NewPenalty(
	userID uuid.UUID,
	weight float64,
	reason string,
	issuedOn time.Time,
	createdAt time.Time,
) (*Penalty, error) {
	return newPenalty(uuid.New(), userID, weight, reason, issuedOn, createdAt, nil)
}

func RestorePenalty(
	id uuid.UUID,
	userID uuid.UUID,
	weight float64,
	reason string,
	issuedOn time.Time,
	createdAt time.Time,
	resolvedAt *time.Time,
) (*Penalty, error) {
	return newPenalty(id, userID, weight, reason, issuedOn, createdAt, resolvedAt)
}

func newPenalty(
	id uuid.UUID,
	userID uuid.UUID,
	weight float64,
	reason string,
	issuedOn time.Time,
	createdAt time.Time,
	resolvedAt *time.Time,
) (*Penalty, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("penalty id is required")
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidResidentID
	}

	normalizedReason, err := normalizeReason(reason)
	if err != nil {
		return nil, err
	}

	normalizedWeight, err := normalizeWeight(weight)
	if err != nil {
		return nil, err
	}

	return &Penalty{
		id:         id,
		userID:     userID,
		weight:     normalizedWeight,
		reason:     normalizedReason,
		issuedOn:   issuedOn,
		createdAt:  createdAt,
		resolvedAt: copyTime(resolvedAt),
	}, nil
}

func normalizeReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	switch {
	case reason == "":
		return "", ErrInvalidReason
	case utf8.RuneCountInString(reason) > maxReasonRuneSize:
		return "", ErrInvalidReason
	default:
		return reason, nil
	}
}

func normalizeWeight(weight float64) (float64, error) {
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		return 0, ErrInvalidWeight
	}

	normalizedWeight := math.Round(weight*10) / 10
	if normalizedWeight < minWeight || normalizedWeight > maxWeight {
		return 0, ErrInvalidWeight
	}

	return normalizedWeight, nil
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}

	copied := *value
	return &copied
}

func (p *Penalty) ID() uuid.UUID        { return p.id }
func (p *Penalty) UserID() uuid.UUID    { return p.userID }
func (p *Penalty) Weight() float64      { return p.weight }
func (p *Penalty) Reason() string       { return p.reason }
func (p *Penalty) IssuedOn() time.Time  { return p.issuedOn }
func (p *Penalty) CreatedAt() time.Time { return p.createdAt }
func (p *Penalty) ResolvedAt() *time.Time {
	return copyTime(p.resolvedAt)
}

func (p *Penalty) IsResolved() bool {
	return p.resolvedAt != nil
}

func (p *Penalty) Resolve(at time.Time) {
	if p.resolvedAt != nil {
		return
	}

	p.resolvedAt = copyTime(&at)
}

type Repository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*Penalty, error)
	Save(ctx context.Context, penalty *Penalty) error
}
