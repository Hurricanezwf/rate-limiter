package limiter

import "github.com/Hurricanezwf/rate-limiter/proto"

// 需要支持并发安全
type LimiterMeta interface {
	// Borrow 申请一次执行资格，如果成功返回nil
	Borrow(cId proto.ClientID, pId proto.RequestPathID) error

	// Return 归还执行资格，如果成功返回nil
	Return(cId proto.ClientID, pId proto.RequestPathID) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	ReturnAll(cId proto.ClientID) error

	// Bytes 将控制器数据序列化，用于持久化
	Bytes() ([]byte, error)

	// LoadFrom 将二进制数据反序列化成该结构
	LoadFrom([]byte) error
}
