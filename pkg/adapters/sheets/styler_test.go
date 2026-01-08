package sheets

import (
	"dorm/pkg/core/ports/dto"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMergeAreaCells(t *testing.T) {
	tasks := []dto.TaskViewModel{
		{ID: uuid.New(), AreaName: "Kitchen"},
		{ID: uuid.New(), AreaName: "Kitchen"},
		{ID: uuid.New(), AreaName: "Kitchen"},
		{ID: uuid.New(), AreaName: "Kitchen"},
		{ID: uuid.New(), AreaName: "Kitchen"},
	}

	requests := MergeAreaCells(12345, tasks)

	assert.Len(t, requests, 1, "Should create one merge request for a single area")

	mergeRequest := requests[0].MergeCells
	assert.NotNil(t, mergeRequest)

	assert.Equal(t, int64(3), mergeRequest.Range.StartRowIndex)
	assert.Equal(t, int64(8), mergeRequest.Range.EndRowIndex)
	assert.Equal(t, int64(0), mergeRequest.Range.StartColumnIndex)
	assert.Equal(t, int64(1), mergeRequest.Range.EndColumnIndex)
	assert.Equal(t, "MERGE_ROWS", mergeRequest.MergeType)
}
