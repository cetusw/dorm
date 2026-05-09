package structure

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRestoreTeam(t *testing.T) {
	id := uuid.New()
	groupID := uuid.New()
	leaderID := uuid.New()
	name := "Test Team"
	color := "#FF0000"
	order := 1

	team := RestoreTeam(id, name, groupID, &leaderID, color, order)

	assert.Equal(t, id, team.ID())
	assert.Equal(t, groupID, team.GroupID())
	assert.Equal(t, &leaderID, team.LeaderID())
	assert.Equal(t, name, team.Name())
	assert.Equal(t, color, team.Color())
	assert.Equal(t, order, team.Order())
}

func TestNewTeam(t *testing.T) {
	groupID := uuid.New()

	team := NewTeam("Test Team", groupID, "#FF0000", 1)

	assert.NotEqual(t, uuid.Nil, team.ID())
	assert.Equal(t, groupID, team.GroupID())
	assert.Nil(t, team.LeaderID())
	assert.Equal(t, "Test Team", team.Name())
	assert.Equal(t, "#FF0000", team.Color())
	assert.Equal(t, 1, team.Order())
}

func TestRestoreGroup(t *testing.T) {
	id := uuid.New()
	leaderID := uuid.New()
	name := "Cleaning Group A"
	sheetID := "spreadsheet-123"
	dormID := int64(10)

	group := RestoreGroup(id, &leaderID, name, sheetID, dormID)

	assert.Equal(t, id, group.ID())
	assert.Equal(t, &leaderID, group.LeaderID())
	assert.Equal(t, name, group.Name())
	assert.Equal(t, sheetID, group.SpreadsheetID())
	assert.Equal(t, dormID, group.DormitoryID())
}

func TestRestoreDormitory(t *testing.T) {
	id := int64(5)
	name := "Main Dorm"
	leaderID := uuid.New()
	city := "Moscow"

	dorm := RestoreDormitory(id, name, &leaderID, city, "ул.", "Ленина", "1")

	assert.Equal(t, id, dorm.ID())
	assert.Equal(t, name, dorm.Name())
	assert.Equal(t, &leaderID, dorm.LeaderID())
	assert.Equal(t, city, dorm.City())
	assert.Equal(t, "Moscow, ул. Ленина, 1", dorm.Address())
}
