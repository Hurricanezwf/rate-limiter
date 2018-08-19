package limiter

import (
	"container/list"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
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
	Borrow(cId ClientID, expire int64) (ResourceID, error)

	// Return 归还执行资格，如果成功返回nil
	Return(cId ClientID, rcId ResourceID) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	ReturnAll(cId ClientID) error

	// Recycle 清理到期未还的资源并且将recycled队列的资源投递到canBorrow队列
	Recycle()

	// 组合Serializer接口
	Serializer
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
	canBorrow *list.List            // 可借的资源队列
	recycled  *list.List            // 已回收的资源队列
	used      map[string]*list.List // 正被使用的资源队列 clientIDHex ==> []borrowRecord
	usedCount int                   // 正被使用的资源数量统计
}

func newLimiterMetaV1(tId ResourceTypeID, quota int) *limiterMetaV1 {
	m := &limiterMetaV1{
		tId:       tId,
		quota:     quota,
		canBorrow: list.New(),
		recycled:  list.New(),
		used:      make(map[string]*list.List),
		usedCount: 0,
	}

	// 初始化等量可借的配额资源
	for i := 0; i < quota; i++ {
		m.canBorrow.PushBack(makeResourceID(tId, i))
	}
	return m
}

// Borrow 借资源
func (m *limiterMetaV1) Borrow(cId ClientID, expire int64) (ResourceID, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 校验额度
	if m.canBorrow.Len() <= 0 {
		return "", ErrQuotaNotEnough
	}

	// 尝试借资源并做记录
	rc := m.canBorrow.Front()
	if rc == nil {
		panic("Exception: Resource is nil")
	} else {
		m.canBorrow.Remove(rc)
	}

	nowTs := time.Now().Unix()
	rcId := rc.Value.(ResourceID)
	record := borrowRecord{
		CID:      cId,
		RCID:     rcId,
		BorrowAt: nowTs,
		ExpireAt: nowTs + expire,
	}

	cIdHex := cId.String()
	link := m.used[cIdHex]
	if link == nil {
		link = list.New()
	}
	link.PushBack(record)
	m.used[cIdHex] = link
	m.usedCount++

	glog.V(1).Infof("Client[%s] borrow %s OK, period:%ds, expireAt:%d", cIdHex, record.RCID, expire, record.ExpireAt)

	return rcId, nil
}

// Return 归还资源
func (m *limiterMetaV1) Return(cId ClientID, rcId ResourceID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 安全性检测
	if m.recycled.Len() >= m.quota {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", m.recycled.Len(), m.quota)
	}

	cIdHex := cId.String()
	link := m.used[cIdHex]
	if link == nil {
		return fmt.Errorf("There's something wrong, no client['%s'] found in used queue", cIdHex)
	}

	// 实施归还
	find := false
	for node := link.Front(); node != nil; node = node.Next() {
		// 找到指定的借出记录
		record := node.Value.(borrowRecord)
		if record.RCID != rcId {
			continue
		}

		// 移出used列表加入到recycled队列
		find = true
		link.Remove(node)
		m.recycled.PushBack(rcId)
		m.usedCount--

		glog.V(3).Infof("Client[%s] return %s to recycle.", cIdHex, rcId)

		// 如果该client没有借入记录，则从map中删除，防止map累积增长
		if link.Len() <= 0 {
			delete(m.used, cIdHex)
		}
		break
	}

	if !find {
		glog.Warningf("No resource[%s] found at used queue. client=%s", rcId, cIdHex)
		return ErrNotFound
	}

	return nil
}

func (m *limiterMetaV1) ReturnAll(cId ClientID) error {
	// TODO:
	return nil
}

// Recycle 检测used队列中的过期资源是否可重用
func (m *limiterMetaV1) Recycle() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 回收过期资源
	nowTs := time.Now().Unix()
	for client, link := range m.used {
		if m.usedCount < 0 {
			panic("There's something wrong, usedCount < 0")
		}
		for record := link.Front(); record != nil; {
			next := record.Next()
			v := record.Value.(borrowRecord)
			if nowTs >= v.ExpireAt {
				// 过期后，从used队列移到canBorrow队列(在过期资源清理和资源重用的周期一致时才可以这样做)
				link.Remove(record)
				m.usedCount--
				m.canBorrow.PushBack(v.RCID)
				glog.V(1).Infof("'%s' borrowed by client['%s'] is expired, force to recycle", v.RCID, client)
			}
			record = next
		}

		// 删除冗余，防止map无限扩张
		if link.Len() <= 0 {
			delete(m.used, client)
		}
	}

	// 资源重用
	if count := m.recycled.Len(); count > 0 {
		m.canBorrow.PushBackList(m.recycled)
		m.recycled.Init()
		glog.V(1).Infof("Refresh %d resources to canBorrow queue because of client's return", count)
	}
}

func (m *limiterMetaV1) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (m *limiterMetaV1) Decode(b []byte) error {
	// TODO:
	return nil
}

func makeResourceID(tId ResourceTypeID, idx int) ResourceID {
	return ResourceID(fmt.Sprintf("%s_rc#%d", tId.String(), idx))
}

func resolveResourceID(rcId ResourceID) (ResourceTypeID, error) {
	arr := strings.Split(string(rcId), "_")
	if len(arr) != 2 {
		return nil, errors.New("Bad ResourceID")
	}

	tId, err := hex.DecodeString(arr[0])
	if err != nil {
		return nil, err
	}
	return ResourceTypeID(tId), nil
}
