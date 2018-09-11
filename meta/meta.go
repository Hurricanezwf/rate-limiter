package meta

func init() {
	RegistBuilder("v2", newMetaV2)
}

// Interface 存储元数据的抽象接口，需要支持并发安全
type Interface interface {
	// RegistQuota 注册资源配额
	// resetInterval: 表示重置资源配额的时间间隔，单位秒
	// timestamp    : 注册时的时间戳
	RegistQuota(rcType []byte, quota uint32, resetInterval, timestamp int64) error

	// Borrow 申请一次执行资格，如果成功返回nil
	// expire   : 表示申请的资源的自动回收时间
	// timestamp: 申请资源时的时间戳
	Borrow(rcType, clientId []byte, expire, timestamp int64) (string, error)

	// Return 归还执行资格，如果成功返回nil
	Return(clientId []byte, rcId string) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	// 返回回收的资源数量
	ReturnAll(rcType, clientId []byte) (uint32, error)

	// Recycle 清理到期未还的资源并且将recycled队列的资源投递到canBorrow队列
	Recycle(timestamp int64)

	// Resources 查询资源的信息
	// 如果rcType为空，将查询全量的资源信息
	ResourceList(rcType []byte) ([]*ResourceDetail, error)

	// Encode 将元数据序列化
	Encode() ([]byte, error)

	// Decode 将元数据反序列化
	Decode(b []byte) error
}

func Default() Interface {
	return New("v2")
}

func New(name string) Interface {
	if builders == nil {
		return nil
	}
	if f := builders[name]; f != nil {
		return f()
	}
	return nil
}

// ResourceDetail 资源详情
type ResourceDetail struct {
	RCType         []byte
	Quota          uint32
	ResetInterval  int64
	CanBorrowCount uint32
	RecycledCount  uint32
	UsedCount      uint32
}
