package individualtask

import (
	"context"
	"github.com/google/uuid"
)

type Repository interface {
	Create(context.Context, *IndividualTask) error
	FindByID(context.Context, uuid.UUID, bool) (*IndividualTask, error)
	Update(context.Context, *IndividualTask, uint64) error
	Complete(context.Context, uuid.UUID, uuid.UUID) (*IndividualTask, error)
	Open(context.Context, uuid.UUID, uuid.UUID) (*IndividualTask, error)
	Reject(context.Context, uuid.UUID) (*IndividualTask, error)
	Verify(context.Context, uuid.UUID) (*IndividualTask, error)
	Delete(context.Context, uuid.UUID) error
}
