package proto

import "errors"

var (
	ErrLeaderNotFound      = errors.New("Leader not found")
	ErrQuotaNotEnough      = errors.New("Quota is not enough")
	ErrExisted             = errors.New("Had been existed")
	ErrResourceNotRegisted = errors.New("Resource didn't regist")
	ErrNoBorrowRecordFound = errors.New("No borrow record found")
	ErrQuotaOverflow       = errors.New("Quota overflow")
)

// Error Code
var (
	ErrCodeBadRequest    byte = 0x01
	ErrCodeServerTooBusy byte = 0x02
)

var (
	ErrBadRequest = errors.New("Bad request")
	ErrTooBusy    = errors.New("Server too busy")
)

var CodeToError = map[byte]error{
	ErrCodeBadRequest:    ErrBadRequest,
	ErrCodeServerTooBusy: ErrTooBusy,
}
