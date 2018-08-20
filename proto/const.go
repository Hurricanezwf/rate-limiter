package proto

// 魔数
var MagicNumber = [1]byte{0x66}

// 快照协议版本
var ProtocolVersion = [1]byte{0x01}

// Raft结点间通信的动作类型
const (
	ActionSnapshot    = ActionType(0xff) // 在所有结点对元数据进行快照处理
	ActionRegistQuota = ActionType(0x01) // 在所有结点注册资源配额
	ActionBorrow      = ActionType(0x02) // 在所有结点进行借资源的操作
	ActionReturn      = ActionType(0x03) // 在所有结点进行归还资源的操作
	ActionDead        = ActionType(0x04)
)

// 值类型
const (
	VTypeUnknown byte = 0x00
	VTypeString  byte = 0x01
	VTypeUint8   byte = 0x02
	VTypeUint32  byte = 0x03
	VTypeUint64  byte = 0x04
	//ValueTypeResourceID     byte = 0x05
	//ValueTypeResourceTypeID byte = 0x06
	//ValueTypeClientID       byte = 0x07
	//ValueTypeBorrowRecord   byte = 0x08
)
