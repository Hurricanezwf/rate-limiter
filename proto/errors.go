package proto

import "errors"

var (
	ErrTooBusy             = errors.New("Too busy")
	ErrLeaderNotFound      = errors.New("Leader not found")
	ErrQuotaNotEnough      = errors.New("Quota is not enough")
	ErrExisted             = errors.New("Had been existed")
	ErrResourceNotRegisted = errors.New("Resource didn't regist")
	ErrNoBorrowRecordFound = errors.New("No borrow record found")
	ErrQuotaOverflow       = errors.New("Quota overflow")
)
