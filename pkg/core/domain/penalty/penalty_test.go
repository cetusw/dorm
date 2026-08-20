package penalty

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPenalty(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.August, 2, 10, 0, 0, 0, time.UTC)

	entity, err := NewPenalty(uuid.New(), EntryTypeIssue, 1.24, " Причина ", createdAt)

	require.NoError(t, err)
	assert.Equal(t, 1.2, entity.Weight())
	assert.Equal(t, "Причина", entity.Reason())
	assert.Equal(t, EntryTypeIssue, entity.Type())
	assert.Equal(t, createdAt, entity.CreatedAt())
}

func TestPenaltyReasonValidation(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.August, 2, 10, 0, 0, 0, time.UTC)
	validReason := strings.Repeat("я", 256)

	testCases := []struct {
		name   string
		reason string
		err    error
	}{
		{name: "trim spaces", reason: "  ok  "},
		{name: "empty reason", reason: "   ", err: ErrInvalidReason},
		{name: "256 runes", reason: validReason},
		{name: "257 runes", reason: validReason + "я", err: ErrInvalidReason},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entity, err := NewPenalty(uuid.New(), EntryTypeIssue, 1.0, testCase.reason, createdAt)
			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(testCase.reason), entity.Reason())
		})
	}
}

func TestPenaltyWeightValidation(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.August, 2, 10, 0, 0, 0, time.UTC)

	testCases := []struct {
		name           string
		weight         float64
		expectedWeight float64
		err            error
	}{
		{name: "round to tenth", weight: 1.26, expectedWeight: 1.3},
		{name: "below minimum after round", weight: 0.04, err: ErrInvalidWeight},
		{name: "minimum", weight: 0.1, expectedWeight: 0.1},
		{name: "maximum", weight: 999999999.9, expectedWeight: 999999999.9},
		{name: "above maximum", weight: 1000000000.0, err: ErrInvalidWeight},
		{name: "nan", weight: math.NaN(), err: ErrInvalidWeight},
		{name: "inf", weight: math.Inf(1), err: ErrInvalidWeight},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			entity, err := NewPenalty(uuid.New(), EntryTypeIssue, testCase.weight, "reason", createdAt)
			if testCase.err != nil {
				assert.ErrorIs(t, err, testCase.err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, testCase.expectedWeight, entity.Weight())
		})
	}
}

func TestPenaltyEntryTypeValidation(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.August, 2, 10, 0, 0, 0, time.UTC)

	_, err := NewPenalty(uuid.New(), EntryType("INVALID"), 1.0, "reason", createdAt)
	assert.ErrorIs(t, err, ErrInvalidPenaltyEntryType)
}
