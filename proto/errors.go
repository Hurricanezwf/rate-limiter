package proto

import "errors"

var (
	ErrLeaderNotFound = errors.New("Leader not found")
	ErrNotLeader      = errors.New("Not leader")
	ErrNoQuotaFound   = errors.New("No quota found")
	ErrQuotaNotEnough = errors.New("Quota is not enough")
	ErrExisted        = errors.New("Had been existed")
	ErrNotFound       = errors.New("Not found")
)
