package meta

import (
	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/golang/protobuf/proto"
)

func init() {
	RegistBuilder("v2", nil)
}

// 需要支持并发安全
type Interface interface {
	// RegistQuota 注册资源配额
	RegistQuota(rcType []byte, quota uint32) error

	// Borrow 申请一次执行资格，如果成功返回nil
	// expire 表示申请的资源的自动回收时间
	Borrow(rcType, clientId []byte, expire int64) (string, error)

	// Return 归还执行资格，如果成功返回nil
	Return(clientId []byte, rcId string) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	// 返回回收的资源数量
	ReturnAll(rcType, clientId []byte) (int, error)

	// Recycle 清理到期未还的资源并且将recycled队列的资源投递到canBorrow队列
	Recycle()

	// Serializer 提供了序列化与反序列化的接口
	encoding.Serializer

	proto.Message
}

func New(name string) Interface {
	// TODO:
	return nil
}

//func New(name string, rcTypeId []byte, quota uint32) Interface {
//	if metaBuilders == nil {
//		return nil
//	}
//	if f := builders[name]; f != nil {
//		return f(rcTypeId, quota)
//	}
//	return nil
//}
//
//func newMetaV2(rcTypeId []byte, quota uint32) Meta {
//	return &MetaV2{
//		RcTypeId:  types.NewBytes(rcTypeId),
//		Quota:     types.NewUint32(quota),
//		CanBorrow: types.NewQueue(),
//		Recycled:  types.NewQueue(),
//		Used:      types.NewMap(),
//	}
//}
//
//func (m *MetaV2) Borrow(clientId []byte, expire int64) (string, error) {
//
//}
//
//func (m *MetaV2) Return(clientId []byte, rcId string) error {
//
//}
//
//func (m *MetaV2) ReturnAll(clientId []byte) (int, error) {
//
//}
//
//func (m *MetaV2) Recycle() {
//
//}
