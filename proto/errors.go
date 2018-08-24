package proto

import "errors"

var (
	ErrTimeout           = errors.New("Timeout")
	ErrTooBusy           = errors.New("Too busy")
	ErrLeaderNotFound    = errors.New("Leader not found")
	ErrNotLeader         = errors.New("Not leader")
	ErrNoQuotaFound      = errors.New("No quota found")
	ErrQuotaNotEnough    = errors.New("Quota is not enough")
	ErrExisted           = errors.New("Had been existed")
	ErrNotFound          = errors.New("Not found")
	ErrMissingResourceID = errors.New("Missing resource id")
)
