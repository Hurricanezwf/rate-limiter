package meta

import (
	"fmt"

	"github.com/golang/protobuf/proto"
)

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

	proto.Message
}

func New(name string, rcTypeId []byte, quota uint32) LimiterMeta {
	if metaBuilders == nil {
		return nil
	}
	if f := metaBuilders[name]; f != nil {
		return f(rcTypeId, quota)
	}
	return nil
}

var metaBuilders = make(map[string]metaBuilder)

type metaBuilder func(rcTypeId []byte, quota uint32) LimiterMeta

func ResgistMetaBuilder(name string, f metaBuilder) {
	if f == nil {
		panic(fmt.Sprintf("MetaBuilder for '%s' is nil", name))
	}
	if _, exist := metaBuilders[name]; exist {
		panic(fmt.Sprintf("MetaBuilder for '%s' had been existed", name))
	}
	metaBuilders[name] = f
}
