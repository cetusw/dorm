package warehouse

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewItemTrimsName(t *testing.T) {
	t.Parallel()

	item, err := NewItem(7, "  Лампа E27  ")

	require.NoError(t, err)
	assert.Equal(t, "Лампа E27", item.Name())
}

func TestNewItemRejectsEmptyName(t *testing.T) {
	t.Parallel()

	_, err := NewItem(7, "   ")

	assert.ErrorIs(t, err, ErrInvalidItemName)
}

func TestNewMovementNormalizesComment(t *testing.T) {
	t.Parallel()

	comment := "  Перегорела  "
	movement, err := NewMovement(uuid.New(), MovementTypeWriteOff, 2, &comment, uuid.New(), time.Now())

	require.NoError(t, err)
	require.NotNil(t, movement.Comment())
	assert.Equal(t, "Перегорела", *movement.Comment())
}

func TestNewMovementRejectsZeroQuantity(t *testing.T) {
	t.Parallel()

	_, err := NewMovement(uuid.New(), MovementTypeAdd, 0, nil, uuid.New(), time.Now())

	assert.ErrorIs(t, err, ErrInvalidMovementQuantity)
}

func TestNewMovementRejectsLongComment(t *testing.T) {
	t.Parallel()

	value := strings.Repeat("а", 257)
	_, err := NewMovement(uuid.New(), MovementTypeAdd, 1, &value, uuid.New(), time.Now())

	assert.ErrorIs(t, err, ErrCommentTooLong)
}
