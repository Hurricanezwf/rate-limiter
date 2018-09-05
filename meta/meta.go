package meta

import (
	"errors"
	fmt "fmt"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/rate-limiter/types"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/golang/protobuf/proto"
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

	// 资源粒度的锁
	mLock map[string]*sync.RWMutex

	// 资源元数据
	m *PB_Meta
}

func newMetaV2() Interface {
	return &metaV2{
		gLock: &sync.RWMutex{},
		mLock: make(map[string]*sync.RWMutex),
		m: &PB_Meta{
			Value: make(map[string]*PB_M),
		},
	}
}

func (m *metaV2) RegistQuota(rcType []byte, quota uint32) error {
	var mLock *sync.RWMutex
	var rcTypeHex = encoding.BytesToStringHex(rcType)

	// 获取资源锁
	m.gLock.Lock()
	mLock = m.mLock[rcTypeHex]
	if mLock == nil {
		mLock = &sync.RWMutex{}
		m.mLock[rcTypeHex] = mLock
	}
	m.gLock.Unlock()

	// 判断资源管理器是否存在
	mLock.Lock()
	defer mLock.Unlock()

	if _, exist := m.m.Value[rcTypeHex]; exist {
		return ErrExisted
	}

	// 注册构造资源管理器
	rcMgr := &PB_M{
		RcTypeId:  rcType,
		Quota:     quota,
		CanBorrow: types.NewQueue(),
		Recycled:  types.NewQueue(),
		Used:      make(map[string]*types.PB_Queue),
		UsedCount: 0,
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
	var mLock *sync.RWMutex
	var rcTypeHex = encoding.BytesToStringHex(rcType)

	// 获取资源锁
	m.gLock.Lock()
	mLock = m.mLock[rcTypeHex]
	if mLock == nil {
		mLock = &sync.RWMutex{}
		m.mLock[rcTypeHex] = mLock
	}
	m.gLock.Unlock()

	// 获取资源管理器
	mLock.Lock()
	defer mLock.Unlock()

	rcMgr, exist := m.m.Value[rcTypeHex]
	if !exist {
		return "", ErrResourceNotRegisted
	}

	// 借资源
	e := rcMgr.CanBorrow.PopFront()
	if e == nil {
		return "", ErrQuotaNotEnough
	}

	rcId, err := types.AnyToString(e.Value)
	if err != nil {
		return "", fmt.Errorf("Resolve resource id failed, %v", err)
	}

	// 构造出借记录
	var clientIdHex = encoding.BytesToStringHex(clientId)
	var nowTs = time.Now().Unix()

	rdQueue := rcMgr.Used[clientIdHex]
	if rdQueue == nil {
		rdQueue = types.NewQueue()
		rcMgr.Used[clientIdHex] = rdQueue
	}

	rcMgr.UsedCount++
	rdQueue.PushBack(&PB_BorrowRecord{
		ClientID: clientId,
		RcID:     rcId.Value,
		BorrowAt: nowTs,
		ExpireAt: nowTs + expire,
	})

	return rcId.Value, nil
}

func (m *metaV2) Return(clientId []byte, rcId string) error {
	rcType, err := ResolveResourceID(rcId)
	if err != nil {
		return fmt.Errorf("Resolve resource id failed, %v", err)
	}

	var mLock *sync.RWMutex
	var rcTypeHex = encoding.BytesToStringHex(rcType)

	// 获取资源锁
	m.gLock.Lock()
	mLock = m.mLock[rcTypeHex]
	if mLock == nil {
		mLock = &sync.RWMutex{}
		m.mLock[rcTypeHex] = mLock
	}
	m.gLock.Unlock()

	// 获取资源管理器
	mLock.Lock()
	defer mLock.Unlock()

	rcMgr, exist := m.m.Value[rcTypeHex]
	if !exist {
		return ErrResourceNotRegisted
	}

	// 安全性检测
	if rcMgr.Recycled.Len() >= rcMgr.Quota {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", rcMgr.Recycled.Len(), rcMgr.Quota)
	}

	// 归还资源
	// 从出借记录中删除并加入recycle队列
	var find = false
	var clientIdHex = encoding.BytesToStringHex(clientId)

	rdList, exist := rcMgr.Used[clientIdHex]
	if !exist || rdList == nil {
		return fmt.Errorf("There's something wrong, no client['%s'] found in used queue", clientIdHex)
	}

	for itr := rdList.Head; itr != nil; itr = itr.Next {
		var record PB_BorrowRecord
		if err := types.UnmarshalAny(itr.Value, &record); err != nil {
			return fmt.Errorf("Unmarshal BorrowRecord failed, %v", err)
		}
		if record.RcID != rcId {
			continue
		}

		//glog.V(3).Infof("Client[%s] return %s to recycle.", clientIdHex, rcId)

		find = true
		rdList.Remove(itr)
		rcMgr.UsedCount--
		rcMgr.Recycled.PushBack(types.NewString(rcId))

		// 如果该client没有借入记录，则从map中删除，防止map累积增长
		if rdList.Len() <= 0 {
			delete(rcMgr.Used, clientIdHex)
		}
		break
	}

	if !find {
		glog.Warningf("No resource[%s] found at used queue. client=%s", rcId, clientIdHex)
		return ErrNotFound
	}

	return nil
}

func (m *metaV2) ReturnAll(rcType, clientId []byte) (uint32, error) {
	var mLock *sync.RWMutex
	var rcTypeHex = encoding.BytesToStringHex(rcType)

	// 获取资源锁
	m.gLock.Lock()
	mLock = m.mLock[rcTypeHex]
	if mLock == nil {
		mLock = &sync.RWMutex{}
		m.mLock[rcTypeHex] = mLock
	}
	m.gLock.Unlock()

	// 获取资源管理器
	mLock.Lock()
	defer mLock.Unlock()

	rcMgr, exist := m.m.Value[rcTypeHex]
	if !exist {
		return 0, ErrResourceNotRegisted
	}

	// 归还指定用户所有指定类型的资源
	clientIdHex := encoding.BytesToStringHex(clientId)
	rdList, exist := rcMgr.Used[clientIdHex]
	if !exist || rdList == nil {
		return 0, nil
	}

	count := rdList.Len()
	for itr := rdList.Head; itr != nil; itr = itr.Next {
		var record PB_BorrowRecord
		if err := types.UnmarshalAny(itr.Value, &record); err != nil {
			return 0, fmt.Errorf("Unmarshal BorrowRecord failed, %v", err)
		}
		if rcMgr.Recycled.Len() >= rcMgr.Quota {
			return 0, errors.New("There's something wrong, recycled queue overflow")
		}
		rcMgr.Recycled.PushBack(types.NewString(record.RcID))
	}
	delete(rcMgr.Used, clientIdHex)
	rcMgr.UsedCount -= count

	return count, nil
}

func (m *metaV2) Recycle() {
	var nowTs int64

	m.gLock.Lock()
	defer m.gLock.Unlock()

	// 扫描所有使用过的资源类型，清理过期资源和恢复资源额度
	// 所有调用过Borrow的资源类型都会有锁存在, 当锁存在但是资源类型不存在时直接从map中删除锁
	for rcTypeHex, mLock := range m.mLock {
		mLock.Lock()

		// 获取资源管理器
		rcMgr, exist := m.m.Value[rcTypeHex]
		if !exist {
			delete(m.mLock, rcTypeHex)
			goto UNLOCK
		}

		// 检测过期资源并回收
		nowTs = time.Now().Unix()
		for clientIdHex, rdList := range rcMgr.Used {
			for itr := rdList.Head; itr != nil; {
				var next = itr.Next
				var record PB_BorrowRecord
				if err := types.UnmarshalAny(itr.Value, &record); err != nil {
					glog.Warning(err.Error())
					goto NEXT
				}
				if record.ExpireAt > nowTs {
					goto NEXT
				}

				// 过期后，从used队列移到canBorrow队列(仅在过期资源清理和资源重用的周期一致时才可以这样做)
				rdList.Remove(itr)
				rcMgr.UsedCount--
				rcMgr.CanBorrow.PushBack(types.NewString(record.RcID))
				glog.V(2).Infof("'%s' borrowed by client['%s'] is expired, force to recycle", record.RcID, clientIdHex)

			NEXT:
				itr = next
			}

			// 当没有此客户的借出记录时，从map中删除
			if rdList.Len() == 0 {
				delete(rcMgr.Used, clientIdHex)
			}
		}

		// 资源重用
		if count := rcMgr.Recycled.Len(); count > 0 {
			rcMgr.CanBorrow.PushBackList(rcMgr.Recycled)
			rcMgr.Recycled.Init()
			glog.V(2).Infof("Refresh %d resources to canBorrow queue because of client's return", count)
		}

	UNLOCK:
		mLock.Unlock()
	}
}

func (m *metaV2) Encode() ([]byte, error) {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	return proto.Marshal(m.m)
}

func (m *metaV2) Decode(b []byte) error {
	if len(b) <= 0 {
		return errors.New("Empty bytes")
	}
	m.gLock.Lock()
	defer m.gLock.Unlock()
	return proto.Unmarshal(b, m.m)
}
