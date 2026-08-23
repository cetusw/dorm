package duty

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDutyParticipantActive(t *testing.T) {
	participant := DutyParticipant{DutyID: uuid.New(), ParticipantID: uuid.New(), Type: ParticipantTypeTemporary}
	assert.True(t, participant.Active())

	now := time.Now()
	participant.ExcludedAt = &now
	assert.False(t, participant.Active())
}
