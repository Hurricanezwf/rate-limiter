package limiter

// 请求路径的ID
type RequestPathID []byte

// 客户端ID
type ClientID []byte

// Limiter 负责请求粒度的访问频率控制
// 需要支持并发安全
type Limiter interface {
	// Get 申请一次执行资格，如果成功返回nil
	Get(cId ClientID, pId RequestPathID) error

	// Put 归还执行资格，如果成功返回nil
	Put(cId ClientID, pId RequestPathID) error

	// PutAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	PutAll(cId ClientID) error

	// Marshal 将控制器数据序列化，用于持久化
	Marshal() ([]byte, error)

	// Unmarshal 将二进制数据反序列化成该结构
	Unmarshal([]byte) error
}
