package proto

import "errors"

var (
	ErrTimeout             = errors.New("Timeout")
	ErrTooBusy             = errors.New("Too busy")
	ErrLeaderNotFound      = errors.New("Leader not found")
	ErrNotLeader           = errors.New("Not leader")
	ErrQuotaNotEnough      = errors.New("Quota is not enough")
	ErrExisted             = errors.New("Had been existed")
	ErrNotFound            = errors.New("Not found")
	ErrResourceNotRegisted = errors.New("Resource didn't regist")
	ErrNoBorrowRecordFound = errors.New("No borrow record found")
	ErrQuotaOverflow       = errors.New("Quota overflow")
)
