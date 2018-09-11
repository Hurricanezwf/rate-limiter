package proto

// 魔数
var MagicNumber byte = 0x66

// 快照协议版本
var ProtocolVersion byte = 0x01

// Raft结点间通信的动作类型
const (
	ActionLeaderNotify byte = 0xa1 // 通知Leader服务地址
	ActionRecycle      byte = 0xa2 // 通知所有结点进行资源清理与重用
	ActionRegistQuota  byte = 0x01 // 在所有结点注册资源配额
	ActionDeleteQuota  byte = 0x02 // 在锁哟结点删除资源配额
	ActionBorrow       byte = 0x03 // 在所有结点进行借资源的操作
	ActionReturn       byte = 0x04 // 在所有结点进行归还资源的操作
	ActionReturnAll    byte = 0x05 // 归还client申请的所有资源
)

// Raft数据存储介质
const (
	RaftStorageMemory = "memory"
	RaftStorageBoltDB = "boltdb"
)

// HTTP请求URI
const (
	RegistQuotaURI  = "/v1/registQuota"
	DeleteQuotaURI  = "/v1/deleteQuota"
	BorrowURI       = "/v1/borrow"
	ReturnURI       = "/v1/return"
	ReturnAllURI    = "/v1/returnAll"
	ResourceListURI = "/v1/rc"
)
