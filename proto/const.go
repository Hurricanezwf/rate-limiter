package proto

// Raft结点间通信的动作类型
const (
	ActionSnapshot    = ActionType(0xff) // 在所有结点对元数据进行快照处理
	ActionRegistQuota = ActionType(0x01) // 在所有结点注册资源配额
	ActionBorrow      = ActionType(0x02) // 在所有结点进行借资源的操作
	ActionReturn      = ActionType(0x03) // 在所有结点进行归还资源的操作
	ActionDead        = ActionType(0x04)
)
