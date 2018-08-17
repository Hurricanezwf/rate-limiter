package proto

const (
	ActionRegistQuota = ActionType(0x01)
	ActionBorrow      = ActionType(0x02)
	ActionReturn      = ActionType(0x03)
	ActionDead        = ActionType(0x04)
)

// 资源类型ID
type ResourceTypeID []byte

// 资源ID
type ResourceID string

// 客户端ID
type ClientID []byte

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
