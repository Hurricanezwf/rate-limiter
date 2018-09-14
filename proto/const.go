package proto

// 快照魔数
var MagicNumber byte = 0x66

// 快照协议版本
var ProtocolVersion byte = 0x01

// 动作类型
const (
	ActionLeaderNotify byte = 0xa1
	ActionRecycle      byte = 0xa2
	ActionRegistQuota  byte = 0x01
	ActionDeleteQuota  byte = 0x02
	ActionBorrow       byte = 0x03
	ActionReturn       byte = 0x04
	ActionReturnAll    byte = 0x05
	ActionResourceList byte = 0x06
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
