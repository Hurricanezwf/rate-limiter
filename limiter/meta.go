package limiter

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

func NewLimiterMeta(tId ResourceTypeID, quota int) LimiterMeta {
	return newLimiterMetaV1(tId, quota)
}

// 需要支持并发安全
type LimiterMeta interface {
	// Borrow 申请一次执行资格，如果成功返回nil
	// expire 表示申请的资源的自动回收时间
	Borrow(tId ResourceTypeID, cId ClientID, expire int64) error

	// Return 归还执行资格，如果成功返回nil
	Return(tId ResourceTypeID, cId ClientID) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	ReturnAll(cId ClientID) error

	// Bytes 将控制器数据序列化，用于持久化
	Bytes() ([]byte, error)

	// LoadFrom 将二进制数据反序列化成该结构
	LoadFrom([]byte) error
}

// borrowRecord 资源借出记录
type borrowRecord struct {
	// ClientID
	CID ClientID

	// 资源ID
	RCID ResourceID

	// 借出时间戳
	BorrowAt int64

	// 到期时间戳
	ExpireAt int64
}

// limiterMetaV1 is an implement of LimiterMeta interface
type limiterMetaV1 struct {
	// 所属的资源类型
	tId ResourceTypeID

	// 拥有的原始资源配额
	quota int

	// 资源调度队列
	mutex     sync.RWMutex
	canBorrow *list.List                // 可借的资源队列
	recycled  *list.List                // 已回收的资源队列
	used      map[string][]borrowRecord // 正被使用的资源队列 clientIDHex ==> []borrowRecord
	usedCount int                       // 正被使用的资源数量统计
}

func newLimiterMetaV1(tId ResourceTypeID, quota int) *limiterMetaV1 {
	m := &limiterMetaV1{
		tId:       tId,
		quota:     quota,
		canBorrow: list.New(),
		recycled:  list.New(),
		used:      make(map[string][]borrowRecord),
		usedCount: 0,
	}

	// 初始化等量可借的配额资源
	for i := 0; i < quota; i++ {
		m.canBorrow.PushBack(makeResourceID(tId, i))
	}
	return m
}

func (m *limiterMetaV1) Borrow(tId ResourceTypeID, cId ClientID, expire int64) error {
	if bytes.Compare(tId, m.tId) != 0 {
		return fmt.Errorf("ResourceType(%s) is not consist with %s", tId, m.tId)
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 校验额度
	if m.usedCount >= m.quota {
		return ErrQuotaNotEnough
	}

	// 尝试借资源并做记录
	rc := m.canBorrow.Front()
	if rc == nil {
		panic("Exception: Resource is nil")
	}

	nowTs := time.Now().Unix()
	record := borrowRecord{
		CID:      cId,
		RCID:     rc.Value.(ResourceID),
		BorrowAt: nowTs,
		ExpireAt: nowTs + expire,
	}

	cIdHex := hex.EncodeToString(cId)
	m.used[cIdHex] = append(m.used[cIdHex], record)
	m.usedCount++

	glog.V(2).Info("Client(%s) borrow %s OK, expireAt %d", cIdHex, record.RCID, record.ExpireAt)

	return nil
}

func (m *limiterMetaV1) Return(tId ResourceTypeID, cId ClientID) error {
	// TODO:
	return nil
}

func (m *limiterMetaV1) ReturnAll(cId ClientID) error {
	// TODO:
	return nil
}

func (m *limiterMetaV1) Bytes() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (m *limiterMetaV1) LoadFrom(b []byte) error {
	// TODO:
	return nil
}

func makeResourceID(tId ResourceTypeID, idx int) ResourceID {
	return ResourceID(fmt.Sprintf("%x_rc#%d", tId, idx))
}
