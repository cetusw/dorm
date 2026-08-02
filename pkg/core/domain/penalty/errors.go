package penalty

import "errors"

var (
	ErrAccessDenied      = errors.New("penalty access denied")
	ErrResidentNotFound  = errors.New("resident not found")
	ErrPenaltyNotFound   = errors.New("penalty not found")
	ErrInvalidResidentID = errors.New("invalid resident id")
	ErrInvalidReason     = errors.New("invalid penalty reason")
	ErrInvalidWeight     = errors.New("invalid penalty weight")
	ErrInvalidIssuedOn   = errors.New("invalid penalty issued date")
	ErrIssuedOnInFuture  = errors.New("penalty issued date is in the future")
)
