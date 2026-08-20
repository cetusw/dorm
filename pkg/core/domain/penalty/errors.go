package penalty

import "errors"

var (
	ErrAccessDenied               = errors.New("penalty access denied")
	ErrPenaltyEntryNotFound       = errors.New("penalty entry not found")
	ErrResidentNotFound           = errors.New("resident not found")
	ErrInvalidResidentID          = errors.New("invalid resident id")
	ErrInvalidReason              = errors.New("invalid penalty reason")
	ErrInvalidWeight              = errors.New("invalid penalty weight")
	ErrInvalidPenaltyEntryType    = errors.New("invalid penalty entry type")
	ErrInsufficientPenaltyBalance = errors.New("penalty balance is insufficient")
	ErrNegativePenaltyHistory     = errors.New("penalty history would become negative")
)
