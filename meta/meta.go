package meta

import (
	"cmp/public-cloud/proxy-layer/logging/glog"
	"errors"
	fmt "fmt"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	pb "github.com/golang/protobuf/proto"
)

func init() {
	RegistBuilder("v2", newMetaV2)
}

// Interface 存储元数据的抽象接口，需要支持并发安全
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
	ReturnAll(rcType, clientId []byte) (uint32, error)

	// Recycle 清理到期未还的资源并且将recycled队列的资源投递到canBorrow队列
	Recycle()

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

// metaV2 是Interface接口的具体实现
type metaV2 struct {
	// 全局锁
	gLock *sync.RWMutex

	// 资源元数据容器
	mgr map[string]*rcManager
}

func newMetaV2() Interface {
	return &metaV2{
		gLock: &sync.RWMutex{},
		mgr:   make(map[string]*rcManager),
	}
}

func (m *metaV2) safeFindManager(rcTypeHex string) *rcManager {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	return m.mgr[rcTypeHex]
}

func (m *metaV2) RegistQuota(rcType []byte, quota uint32) error {
	m.gLock.Lock()
	defer m.gLock.Unlock()

	// 验证是否已经注册
	rcTypeHex := encoding.BytesToStringHex(rcType)
	rcMgr := m.mgr[rcTypeHex]
	if rcMgr != nil {
		return ErrExisted
	}

	// 初始化资源元数据
	rcMgr = newRCManager(rcType, quota)
	for i := uint32(0); i < quota; i++ {
		rcMgr.canBorrow.PushBack(MakeResourceID(rcType, i))
	}

	m.mgr[rcTypeHex] = rcMgr

	return nil
}

func (m *metaV2) Borrow(rcType, clientId []byte, expire int64) (string, error) {
	rcTypeHex := encoding.BytesToStringHex(rcType)
	rcMgr := m.safeFindManager(rcTypeHex)
	if rcMgr == nil {
		return "", ErrResourceNotRegisted
	}
	return rcMgr.safeBorrow(rcType, clientId, expire)
}

func (m *metaV2) Return(clientId []byte, rcId string) error {
	rcType, err := ResolveResourceID(rcId)
	if err != nil {
		return fmt.Errorf("Resolve resource id failed, %v", err)
	}

	rcTypeHex := encoding.BytesToStringHex(rcType)
	rcMgr := m.safeFindManager(rcTypeHex)
	if rcMgr == nil {
		return ErrResourceNotRegisted
	}
	return rcMgr.safeReturn(clientId, rcId)
}

func (m *metaV2) ReturnAll(rcType, clientId []byte) (uint32, error) {
	rcTypeHex := encoding.BytesToStringHex(rcType)
	rcMgr := m.safeFindManager(rcTypeHex)
	if rcMgr == nil {
		return 0, ErrResourceNotRegisted
	}
	return rcMgr.safeReturnAll(clientId)
}

func (m *metaV2) Recycle() {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	for _, rcMgr := range m.mgr {
		rcMgr.safeRecycle()
	}
}

func (m *metaV2) Encode() ([]byte, error) {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	return pb.Marshal(m.copyToProtobuf())
}

func (m *metaV2) Decode(b []byte) error {
	if len(b) <= 0 {
		return errors.New("Empty bytes")
	}

	pbMeta := &PB_Meta{}
	err := pb.Unmarshal(b, pbMeta)
	if err == nil {
		m.copyFromProtobuf(pbMeta)
	}
	return err
}

func (m *metaV2) copyToProtobuf() *PB_Meta {
	m.gLock.RLock()
	defer m.gLock.RUnlock()

	pb := &PB_Meta{
		Value: make(map[string]*PB_Manager, len(m.mgr)),
	}

	for rcTypeHex, rcMgr := range m.mgr {
		pbMgr := &PB_Manager{
			RCType:    rcMgr.rcType,
			Quota:     rcMgr.quota,
			CanBorrow: make([]string, 0, rcMgr.canBorrow.Len()),
			Recycled:  make([]string, 0, rcMgr.recycled.Len()),
			Used:      make([]*PB_BorrowRecord, 0, rcMgr.usedCount),
			UsedCount: rcMgr.usedCount,
		}
		pb.Value[rcTypeHex] = pbMgr

		for itr := rcMgr.canBorrow.Front(); itr != nil; itr = itr.Next() {
			rcId := itr.Value.(string)
			pbMgr.CanBorrow = append(pbMgr.CanBorrow, rcId)
		}
		for itr := rcMgr.recycled.Front(); itr != nil; itr = itr.Next() {
			rcId := itr.Value.(string)
			pbMgr.Recycled = append(pbMgr.Recycled, rcId)
		}
		for clientIdHex, rdTable := range rcMgr.used {
			clientId, err := encoding.StringHexToBytes(clientIdHex)
			if err != nil {
				glog.Warningf("Convert clientIdHex '%s' to bytes failed, %v", clientIdHex, err)
				continue
			}
			for rcId, rd := range rdTable {
				pbMgr.Used = append(pbMgr.Used, &PB_BorrowRecord{
					ClientID: clientId,
					RCID:     rcId,
					BorrowAt: rd.borrowAt,
					ExpireAt: rd.expireAt,
				})
			}
		}
	}

	return pb
}

func (m *metaV2) copyFromProtobuf(pb *PB_Meta) {
	m.gLock.Lock()
	defer m.gLock.Unlock()

	for rcTypeHex, pbMgr := range pb.Value {
		rcMgr := newRCManager(pbMgr.RCType, pbMgr.Quota)
		rcMgr.quota = pbMgr.Quota
		rcMgr.usedCount = pbMgr.UsedCount

		for _, rcId := range pbMgr.CanBorrow {
			rcMgr.canBorrow.PushBack(rcId)
		}
		for _, rcId := range pbMgr.Recycled {
			rcMgr.recycled.PushBack(rcId)
		}
		for _, pbRd := range pbMgr.Used {
			clientIdHex := encoding.BytesToStringHex(pbRd.ClientID)
			rdTable := rcMgr.used[clientIdHex]
			if rdTable == nil {
				rdTable = make(map[string]*borrowRecord)
			}
			rdTable[pbRd.RCID] = &borrowRecord{
				borrowAt: pbRd.BorrowAt,
				expireAt: pbRd.ExpireAt,
			}
			rcMgr.used[clientIdHex] = rdTable
		}
		m.mgr[rcTypeHex] = rcMgr
	}
}
