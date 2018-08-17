package proto

import "errors"

var (
	ErrNotLeader      = errors.New("Not leader")
	ErrNoQuotaFound   = errors.New("No quota found")
	ErrQuotaNotEnough = errors.New("Quota is not enough")
	ErrExisted        = errors.New("Had been existed")
)
