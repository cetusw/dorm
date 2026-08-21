package individualtask

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *IndividualTask) error
	FindByID(context.Context, uuid.UUID, bool) (*IndividualTask, error)
	Update(context.Context, *IndividualTask, uint64, bool) error
	Complete(context.Context, uuid.UUID, uuid.UUID, int64) (*IndividualTask, error)
	Reject(context.Context, uuid.UUID, int64) (*IndividualTask, error)
	Verify(context.Context, uuid.UUID, int64) (*IndividualTask, error)
	Delete(context.Context, uuid.UUID, int64) error
}
