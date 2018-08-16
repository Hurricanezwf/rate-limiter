package proto

const (
	ActionBorrow = ActionType(0x01)
	ActionReturn = ActionType(0x02)
	ActionDead   = ActionType(0x03)
)

// 请求路径的ID
type RequestPathID []byte

// 客户端ID
type ClientID []byte

// Action类型
type ActionType uint8

type Request struct {
	Action ActionType    `json:"action,omitempty"`
	CID    ClientID      `json:"cId,omitempty"`
	PID    RequestPathID `json:"pId,omitempty"`
}
