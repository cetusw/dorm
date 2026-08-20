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
	id        uuid.UUID
	userID    uuid.UUID
	entryType EntryType
	weight    float64
	reason    string
	createdAt time.Time
}

type EntryType string

const (
	EntryTypeIssue   EntryType = "ISSUE"
	EntryTypeResolve EntryType = "RESOLVE"
)

func NewPenalty(
	userID uuid.UUID,
	entryType EntryType,
	weight float64,
	reason string,
	createdAt time.Time,
) (*Penalty, error) {
	return newPenalty(uuid.New(), userID, entryType, weight, reason, createdAt)
}

func RestorePenalty(
	id uuid.UUID,
	userID uuid.UUID,
	entryType EntryType,
	weight float64,
	reason string,
	createdAt time.Time,
) (*Penalty, error) {
	return newPenalty(id, userID, entryType, weight, reason, createdAt)
}

func newPenalty(
	id uuid.UUID,
	userID uuid.UUID,
	entryType EntryType,
	weight float64,
	reason string,
	createdAt time.Time,
) (*Penalty, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("penalty id is required")
	}
	if userID == uuid.Nil {
		return nil, ErrInvalidResidentID
	}
	if !entryType.IsValid() {
		return nil, ErrInvalidPenaltyEntryType
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
		id:        id,
		userID:    userID,
		entryType: entryType,
		weight:    normalizedWeight,
		reason:    normalizedReason,
		createdAt: createdAt,
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

func (t EntryType) IsValid() bool {
	return t == EntryTypeIssue || t == EntryTypeResolve
}

func (p *Penalty) ID() uuid.UUID        { return p.id }
func (p *Penalty) UserID() uuid.UUID    { return p.userID }
func (p *Penalty) Type() EntryType      { return p.entryType }
func (p *Penalty) Weight() float64      { return p.weight }
func (p *Penalty) Reason() string       { return p.reason }
func (p *Penalty) CreatedAt() time.Time { return p.createdAt }

type Repository interface {
	Create(ctx context.Context, penalty *Penalty) error
	CreateResolve(ctx context.Context, penalty *Penalty) error
	FindByID(ctx context.Context, id uuid.UUID) (*Penalty, error)
	Update(ctx context.Context, penalty *Penalty) error
	Delete(ctx context.Context, id uuid.UUID) error
}
