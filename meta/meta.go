package meta

import (
	fmt "fmt"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/Hurricanezwf/rate-limiter/types"
)

func init() {
	RegistBuilder("v2", newMetaV2)
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

	// Encode 将元数据序列化
	Encode() ([]byte, error)

	// Decode 将元数据反序列化
	Decode(b []byte) error
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

// metaV2 是Interface接口的具体实现
type metaV2 struct {
	gLock *sync.RWMutex
	m     *PB_Meta
}

func newMetaV2() Interface {
	return &metaV2{
		gLock: &sync.RWMutex{},
		m: &PB_Meta{
			Value: make(map[string]*PB_M),
		},
	}
}

func (m *metaV2) RegistQuota(rcType []byte, quota uint32) error {
	m.gLock.Lock()
	defer m.gLock.Unlock()

	rcTypeHex := encoding.BytesToStringHex(rcType)
	if _, exist := m.m.Value[rcTypeHex]; exist {
		return fmt.Errorf("Resource '%s' had been existed", rcTypeHex)
	}

	rcMgr := &PB_M{
		RcTypeId:  types.NewBytes(rcType),
		Quota:     types.NewUint32(quota),
		CanBorrow: types.NewQueue(),
	}
	for i := uint32(0); i < quota; i++ {
		rcId := types.NewString(MakeResourceID(rcType, i))
		if _, err := rcMgr.CanBorrow.PushBack(rcId); err != nil {
			return fmt.Errorf("Generate canBorrow resource failed, %v", err)
		}
	}

	m.m.Value[rcTypeHex] = rcMgr
	return nil
}

func (m *metaV2) Borrow(rcType, clientId []byte, expire int64) (string, error) {
	// TODO:
	return "", nil
}

func (m *metaV2) Return(clientId []byte, rcId string) error {
	// TODO:
	return nil
}

func (m *metaV2) ReturnAll(rcType, clientId []byte) (int, error) {
	// TODO:
	return 0, nil
}

func (m *metaV2) Recycle() {
	// TODO:
}

func (m *metaV2) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (m *metaV2) Decode(b []byte) error {
	// TODO:
	return nil
}
