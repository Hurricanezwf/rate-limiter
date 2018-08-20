package limiter

import (
	"fmt"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoder"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

func NewLimiterMeta(rcTypeId []byte, quota uint32) LimiterMeta {
	return newLimiterMetaV1(rcTypeId, quota)
}

func NewLimiterMetaFromBytes(b []byte) (LimiterMeta, error) {
	var m limiterMetaV1
	if _, err := m.Decode(b); err != nil {
		return nil, err
	}
	return &m, nil
}

// 需要支持并发安全
type LimiterMeta interface {
	// Borrow 申请一次执行资格，如果成功返回nil
	// expire 表示申请的资源的自动回收时间
	Borrow(clientId []byte, expire int64) (string, error)

	// Return 归还执行资格，如果成功返回nil
	Return(clientId []byte, rcId string) error

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	ReturnAll(clientId []byte) error

	// Recycle 清理到期未还的资源并且将recycled队列的资源投递到canBorrow队列
	Recycle()

	// 组合Serializer接口
	Serializer
}

// limiterMetaV1 is an implement of LimiterMeta interface
type limiterMetaV1 struct {
	// 所属的资源类型
	rcTypeId []byte

	// 拥有的原始资源配额
	quota uint32

	// 资源调度队列
	mutex     sync.RWMutex
	canBorrow *Queue            // 可借的资源队列
	recycled  *Queue            // 已回收的资源队列
	used      map[string]*Queue // 正被使用的资源队列 clientIDHex ==> borrowRecord queue
	usedCount uint32            // 正被使用的资源数量统计
}

func newLimiterMetaV1(rcTypeId []byte, quota uint32) *limiterMetaV1 {
	m := &limiterMetaV1{
		rcTypeId:  rcTypeId,
		quota:     quota,
		canBorrow: NewQueue(),
		recycled:  NewQueue(),
		used:      make(map[string]*Queue),
		usedCount: uint32(0),
	}

	// 初始化等量可借的配额资源
	for i := uint32(0); i < quota; i++ {
		m.canBorrow.PushBack(MakeResourceID(rcTypeId, i))
	}
	return m
}

// Borrow 借资源
func (m *limiterMetaV1) Borrow(clientId []byte, expire int64) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 校验额度
	if m.canBorrow.Len() <= 0 {
		return "", ErrQuotaNotEnough
	}

	// 尝试借资源并做记录
	rcIdInterface, ok := m.canBorrow.PopFront()
	if !ok {
		panic("Exception: No resource found")
	}
	if rcIdInterface == nil {
		panic("Exception: Nil resource found")
	}

	rcId := rcIdInterface.(string)
	nowTs := time.Now().Unix()
	record := borrowRecord{
		ClientID: clientId,
		RCID:     rcId,
		BorrowAt: nowTs,
		ExpireAt: nowTs + expire,
	}

	clientIdHex := encoder.BytesToStringHex(clientId)
	queue := m.used[clientIdHex]
	if queue == nil {
		queue = NewQueue()
	}
	queue.PushBack(record)
	m.used[clientIdHex] = queue
	m.usedCount++

	glog.V(1).Infof("Client[%s] borrow %s OK, period:%ds, expireAt:%d", clientIdHex, record.RCID, expire, record.ExpireAt)

	return rcId, nil
}

// Return 归还资源
func (m *limiterMetaV1) Return(clientId []byte, rcId string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 安全性检测
	if uint32(m.recycled.Len()) >= m.quota {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", m.recycled.Len(), m.quota)
	}

	clientIdHex := encoder.BytesToStringHex(clientId)
	queue := m.used[clientIdHex]
	if queue == nil {
		return fmt.Errorf("There's something wrong, no client['%s'] found in used queue", clientIdHex)
	}

	// 实施归还
	find := false
	for node := queue.Front(); node != nil; node = node.Next() {
		// 找到指定的借出记录
		record := node.Value.(borrowRecord)
		if record.RCID != rclientId {
			continue
		}

		// 移出used列表加入到recycled队列
		find = true
		queue.Remove(node)
		m.recycled.PushBack(rcId)
		m.usedCount--

		glog.V(3).Infof("Client[%s] return %s to recycle.", clientIdHex, rcId)

		// 如果该client没有借入记录，则从map中删除，防止map累积增长
		if queue.Len() <= 0 {
			delete(m.used, clientIdHex)
		}
		break
	}

	if !find {
		glog.Warningf("No resource[%s] found at used queue. client=%s", rcId, clientIdHex)
		return ErrNotFound
	}

	return nil
}

func (m *limiterMetaV1) ReturnAll(clientId []byte) error {
	// TODO:
	return nil
}

// Recycle 检测used队列中的过期资源是否可重用
func (m *limiterMetaV1) Recycle() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 回收过期资源
	nowTs := time.Now().Unix()
	for client, queue := range m.used {
		if m.usedCount < 0 {
			panic("There's something wrong, usedCount < 0")
		}
		for record := queue.Front(); record != nil; {
			next := record.Next()
			v := record.Value.(borrowRecord)
			if nowTs >= v.ExpireAt {
				// 过期后，从used队列移到canBorrow队列(在过期资源清理和资源重用的周期一致时才可以这样做)
				queue.Remove(record)
				m.usedCount--
				m.canBorrow.PushBack(v.RCID)
				glog.V(1).Infof("'%s' borrowed by client['%s'] is expired, force to recycle", v.RCID, client)
			}
			record = next
		}

		// 删除冗余，防止map无限扩张
		if queue.Len() <= 0 {
			delete(m.used, client)
		}
	}

	// 资源重用
	if count := m.recycled.Len(); count > 0 {
		m.canBorrow.PushBackQueue(m.recycled)
		m.recycled.Init()
		glog.V(1).Infof("Refresh %d resources to canBorrow queue because of client's return", count)
	}
}

func (m *limiterMetaV1) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (m *limiterMetaV1) Decode(b []byte) ([]byte, error) {
	// TODO:
	return b, nil
}

// borrowRecord 资源借出记录
type borrowRecord struct {
	// 客户端ID
	ClientID []byte

	// 资源ID
	RCID string

	// 借出时间戳
	BorrowAt int64

	// 到期时间戳
	ExpireAt int64
}

func (rd *borrowRecord) Encode() ([]byte, error) {
	// TODO
	return nil, nil
}

func (rd *borrowRecord) Decode(b []byte) ([]byte, error) {
	// TODO
	return nil, nil
}
