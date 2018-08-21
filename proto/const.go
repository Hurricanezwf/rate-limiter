package proto

// 魔数
var MagicNumber byte = 0x66

// 快照协议版本
var ProtocolVersion byte = 0x01

// Raft结点间通信的动作类型
const (
	ActionRegistQuota byte = 0x01 // 在所有结点注册资源配额
	ActionBorrow      byte = 0x02 // 在所有结点进行借资源的操作
	ActionReturn      byte = 0x03 // 在所有结点进行归还资源的操作
	ActionDead        byte = 0x04
)
