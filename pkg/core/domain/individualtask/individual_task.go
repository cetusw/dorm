package individualtask

import (
	"errors"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const MaxWeight = 999999999.9

var (
	ErrNotFound          = errors.New("individual task not found")
	ErrAccessDenied      = errors.New("individual task access denied")
	ErrInvalidTitle      = errors.New("invalid individual task title")
	ErrInvalidWeight     = errors.New("invalid individual task redemption weight")
	ErrInvalidDeadline   = errors.New("invalid individual task deadline")
	ErrInvalidArea       = errors.New("invalid individual task area")
	ErrInvalidResident   = errors.New("invalid individual task resident")
	ErrInvalidTransition = errors.New("invalid individual task transition")
	ErrVersionConflict   = errors.New("individual task version conflict")
	ErrCapacityExceeded  = errors.New("individual task redemption capacity exceeded")
)

type Status string

const (
	StatusIssued    Status = "ISSUED"
	StatusCompleted Status = "COMPLETED"
	StatusVerified  Status = "VERIFIED"
)

type IndividualTask struct {
	ID               uuid.UUID
	DormitoryID      int64
	ResidentID       uuid.UUID
	AreaID           *int
	Title            string
	RedemptionWeight float64
	Status           Status
	Deadline         *time.Time
	CompletedAt      *time.Time
	VerifiedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DeletedAt        *time.Time
	Version          uint64
}

func New(dormitoryID int64, residentID uuid.UUID, areaID *int, title string, weight float64, deadline *time.Time, now time.Time) (*IndividualTask, error) {
	title, err := NormalizeTitle(title)
	if err != nil {
		return nil, err
	}
	weight, err = NormalizeWeight(weight)
	if err != nil {
		return nil, err
	}
	if dormitoryID <= 0 || residentID == uuid.Nil {
		return nil, ErrInvalidResident
	}
	return &IndividualTask{ID: uuid.New(), DormitoryID: dormitoryID, ResidentID: residentID, AreaID: areaID, Title: title, RedemptionWeight: weight, Status: StatusIssued, Deadline: deadline, CreatedAt: now, UpdatedAt: now, Version: 1}, nil
}

func NormalizeTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" || utf8.RuneCountInString(title) > 255 {
		return "", ErrInvalidTitle
	}
	return title, nil
}

func NormalizeWeight(weight float64) (float64, error) {
	if math.IsNaN(weight) || math.IsInf(weight, 0) {
		return 0, ErrInvalidWeight
	}
	weight = math.Round(weight*10) / 10
	if weight < 0 || weight > MaxWeight {
		return 0, ErrInvalidWeight
	}
	return weight, nil
}

func (t *IndividualTask) Complete(at time.Time) error {
	switch t.Status {
	case StatusIssued:
		t.Status = StatusCompleted
		t.CompletedAt = &at
		t.touch(at)
		return nil
	case StatusCompleted:
		return nil
	default:
		return ErrInvalidTransition
	}
}
func (t *IndividualTask) Reject(at time.Time) error {
	switch t.Status {
	case StatusCompleted:
		t.Status = StatusIssued
		t.CompletedAt = nil
		t.touch(at)
		return nil
	case StatusIssued:
		return nil
	default:
		return ErrInvalidTransition
	}
}
func (t *IndividualTask) Verify(at time.Time) error {
	switch t.Status {
	case StatusCompleted:
		t.Status = StatusVerified
		t.VerifiedAt = &at
		t.touch(at)
		return nil
	case StatusVerified:
		return nil
	default:
		return ErrInvalidTransition
	}
}
func (t *IndividualTask) Delete(at time.Time) error {
	if t.DeletedAt != nil {
		return nil
	}
	if t.Status == StatusVerified {
		return ErrInvalidTransition
	}
	t.DeletedAt = &at
	t.touch(at)
	return nil
}
func (t *IndividualTask) CanEdit() bool      { return t.DeletedAt == nil && t.Status == StatusIssued }
func (t *IndividualTask) touch(at time.Time) { t.UpdatedAt = at; t.Version++ }
