package proto

import "encoding/hex"

// 资源类型ID
type ResourceTypeID []byte

func (id ResourceTypeID) String() string {
	return hex.EncodeToString(id)
}

// 资源ID
type ResourceID string

// 客户端ID
type ClientID []byte

func (id ClientID) String() string {
	return hex.EncodeToString(id)
}

// Action类型
type ActionType uint8

type Request struct {
	Action ActionType     `json:"action,omitempty"`
	TID    ResourceTypeID `json:"tId,omitempty"`
	RCID   ResourceID     `json:"rcId,omitempty"`
	CID    ClientID       `json:"cId,omitempty"`
	Expire int64          `json:"expire,omitempty"`
	Quota  int            `json:"quota,omitempty"`
}

type Response struct {
	Err    error
	Result interface{}
}
