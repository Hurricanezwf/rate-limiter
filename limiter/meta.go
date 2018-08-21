package limiter

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
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
	encoding.Serializer
}

// limiterMetaV1 is an implement of LimiterMeta interface
type limiterMetaV1 struct {
	mutex sync.RWMutex

	// 所属的资源类型
	rcTypeId *encoding.Bytes

	// 拥有的原始资源配额
	quota *encoding.Uint32

	// 资源调度队列
	canBorrow *encoding.Queue  // 可借的资源队列
	recycled  *encoding.Queue  // 已回收的资源队列
	used      *encoding.Map    // 已借出资源的记录队列. clientIdHex ==> *BorrowRecord queue
	usedCount *encoding.Uint32 // 正被使用的资源数量统计
}

func newLimiterMetaV1(rcTypeId []byte, quota uint32) *limiterMetaV1 {
	m := &limiterMetaV1{
		rcTypeId:  encoding.NewBytes(rcTypeId),
		quota:     encoding.NewUint32(quota),
		canBorrow: encoding.NewQueue(),
		recycled:  encoding.NewQueue(),
		used:      encoding.NewMap(),
		usedCount: encoding.NewUint32(uint32(0)),
	}

	// 初始化等量可借的配额资源
	for i := uint32(0); i < quota; i++ {
		rcId := MakeResourceID(rcTypeId, i)
		m.canBorrow.PushBack(encoding.NewString(rcId))
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

	// 从可借队列中取出资源
	rcIdInterface, ok := m.canBorrow.PopFront()
	if !ok {
		panic("Exception: No resource found")
	}
	if rcIdInterface == nil {
		panic("Exception: Nil resource found")
	}

	// 构造出借记录
	rcId := rcIdInterface.(*encoding.String)
	nowTs := time.Now().Unix()
	record := &BorrowRecord{
		ClientID: encoding.NewBytes(clientId),
		RCID:     rcId,
		BorrowAt: encoding.NewInt64(nowTs),
		ExpireAt: encoding.NewInt64(nowTs + expire),
	}

	// 添加到出借记录列表中
	var q *encoding.Queue
	var clientIdHex = record.ClientID.Hex()

	queue, exist := m.used.Get(clientIdHex)
	if !exist || queue == nil {
		q = encoding.NewQueue()
	} else {
		q = queue.(*encoding.Queue)
	}

	q.PushBack(record)
	m.used.Set(clientIdHex, q)
	m.usedCount.Incr(uint32(1))

	glog.V(1).Infof("Client[%s] borrow %s OK, period:%ds, expireAt:%d", clientIdHex, record.RCID.Value(), expire, record.ExpireAt.Value())

	return rcId.Value(), nil
}

// Return 归还资源
func (m *limiterMetaV1) Return(clientId []byte, rcId string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 安全性检测
	if uint32(m.recycled.Len()) >= m.quota.Value() {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", m.recycled.Len(), m.quota.Value())
	}

	clientIdHex := encoding.BytesToStringHex(clientId)
	q, exist := m.used.Get(clientIdHex)
	if !exist || q == nil {
		return fmt.Errorf("There's something wrong, no client['%s'] found in used queue", clientIdHex)
	}

	// 实施归还
	queue := q.(*encoding.Queue)
	find := false
	for node := queue.Front(); node.IsNil() == false; node = node.Next() {
		// 找到指定的借出记录
		record := node.Value().(*BorrowRecord)
		if record.RCID.Value() != rcId {
			continue
		}

		// 移出used列表加入到recycled队列
		find = true
		queue.Remove(node)
		m.recycled.PushBack(encoding.NewString(rcId))
		m.usedCount.Decr(uint32(1))

		glog.V(3).Infof("Client[%s] return %s to recycle.", clientIdHex, rcId)

		// 如果该client没有借入记录，则从map中删除，防止map累积增长
		if queue.Len() <= 0 {
			m.used.Delete(clientIdHex)
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

	quitC := make(chan struct{})
	defer close(quitC)

	for pair := range m.used.Range(quitC) {
		client := pair.K
		queue := pair.V.(*encoding.Queue)

		if m.usedCount.Value() < 0 {
			panic("There's something wrong, usedCount < 0")
		}
		for record := queue.Front(); record.IsNil() == false; {
			next := record.Next()
			v := record.Value().(*BorrowRecord)
			if nowTs >= v.ExpireAt.Value() {
				// 过期后，从used队列移到canBorrow队列(在过期资源清理和资源重用的周期一致时才可以这样做)
				queue.Remove(record)
				m.usedCount.Decr(uint32(1))
				m.canBorrow.PushBack(v.RCID)
				glog.V(1).Infof("'%s' borrowed by client['%s'] is expired, force to recycle", v.RCID.Value(), client)
			}
			record = next
		}

		// 删除冗余，防止map无限扩张
		if queue.Len() <= 0 {
			m.used.Delete(client)
		}
	}

	// 资源重用
	if count := m.recycled.Len(); count > 0 {
		m.canBorrow.PushBackQueue(m.recycled)
		m.recycled.Init()
		glog.V(1).Infof("Refresh %d resources to canBorrow queue because of client's return", count)
	}
}

// Encode encode limiterMetaV1 to format which used to binary persist
// Encode Format:
// > [1 byte]  data type
// > [8 bytes] data len
// > [N bytes] data content
func (m *limiterMetaV1) Encode() ([]byte, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// encode rcTypeId
	rcTypeIdBytes, err := m.rcTypeId.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode limiterMetaV1.rcTypeIdBytes failed, %v", err)
	}

	// encode quota
	quotaBytes, err := m.quota.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode LimiterMeta.quota failed, %v", err)
	}

	// encode canBorrow
	canBorrowBytes, err := m.canBorrow.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode LimiterMeta.canBorrow failed, %v", err)
	}

	// encode recycled
	recycledBytes, err := m.recycled.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode LimiterMeta.recycled failed, %v", err)
	}

	// encode used
	usedBytes, err := m.used.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode LimiterMeta.used failed, %v", err)
	}

	// encode usedCount
	usedCountBytes, err := m.usedCount.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode LimiterMeta.usedCountBytes failed, %v", err)
	}

	// calc bytes count
	var bytesCount uint64
	bytesCount += uint64(len(rcTypeIdBytes))
	bytesCount += uint64(len(quotaBytes))
	bytesCount += uint64(len(canBorrowBytes))
	bytesCount += uint64(len(recycledBytes))
	bytesCount += uint64(len(usedBytes))
	bytesCount += uint64(len(usedCountBytes))

	vLen := make([]byte, 8)
	binary.BigEndian.PutUint64(vLen[0:8], uint64(bytesCount))

	// build
	buf := bytes.NewBuffer(make([]byte, 0, 9+bytesCount))
	buf.WriteByte(encoding.VTypeLimiterMeta)
	buf.Write(vLen)
	buf.Write(rcTypeIdBytes)
	buf.Write(quotaBytes)
	buf.Write(canBorrowBytes)
	buf.Write(recycledBytes)
	buf.Write(usedBytes)
	buf.Write(usedCountBytes)

	return buf.Bytes(), nil
}

func (m *limiterMetaV1) Decode(b []byte) ([]byte, error) {
	// 校验头部
	if len(b) < 9 {
		return nil, errors.New("Bad encoded format for LimiterMeta, too short")
	}

	// 校验数据类型
	if vType := b[0]; vType != encoding.VTypeLimiterMeta {
		return nil, fmt.Errorf("Bad encoded format for LimiterMeta, VType(%#x) don't match %#x", vType, encoding.VTypeLimiterMeta)
	}

	// 校验数据长度
	vLen := binary.BigEndian.Uint64(b[1:9])
	if uint64(len(b)) < 9+vLen {
		return nil, errors.New("Bad encoded format for LimiterMeta, too short")
	}

	var err error

	// 解析rcTypeId
	b = b[9:]
	rcTypeId := encoding.NewBytes(nil)
	b, err = rcTypeId.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.rcTypeId failed, %v", err)
	}

	// 解析quota
	quota := encoding.NewUint32(0)
	b, err = quota.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.quota failed, %v", err)
	}

	// 解析canBorrow
	canBorrow := encoding.NewQueue()
	b, err = canBorrow.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.canBorrow failed, %v", err)
	}

	// 解析recycled
	recycled := encoding.NewQueue()
	b, err = recycled.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.recycled failed, %v", err)
	}

	// 解析used
	used := encoding.NewMap()
	b, err = used.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.used failed, %v", err)
	}

	// 解析usedCount
	usedCount := encoding.NewUint32(0)
	b, err = usedCount.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode LimiterMeta.usedCount failed, %v", err)
	}

	m.mutex.Lock()
	m.rcTypeId = rcTypeId
	m.quota = quota
	m.canBorrow = canBorrow
	m.recycled = recycled
	m.used = used
	m.usedCount = usedCount
	m.mutex.Unlock()

	return b, nil
}

// BorrowRecord 资源借出记录
type BorrowRecord struct {
	// 客户端ID
	ClientID *encoding.Bytes

	// 资源ID
	RCID *encoding.String

	// 借出时间戳
	BorrowAt *encoding.Int64

	// 到期时间戳
	ExpireAt *encoding.Int64
}

// Encode encode BorrowRecord object to format which used to binary persist
// Encode Format:
// > [1 byte]  data type
// > [8 bytes] data len
// > [N bytes] data content
func (rd *BorrowRecord) Encode() ([]byte, error) {
	// encode ClientID
	clientIdBytes, err := rd.ClientID.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode BorrowRecord.ClientID failed, %v", err)
	}

	// encode RCID
	rcIdBytes, err := rd.RCID.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode BorrowRecordRCID failed, %v", err)
	}

	// encode BorrowAt
	borrowAtBytes, err := rd.BorrowAt.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode BorrowRecord.BorrowAt failed, %v", err)
	}

	// encode ExpireAt
	expireAtBytes, err := rd.ExpireAt.Encode()
	if err != nil {
		return nil, fmt.Errorf("Encode BorrowRecord.ExpireAt failed, %v", err)
	}

	// calc bytes count
	var memberBytesCount uint64
	memberBytesCount += uint64(len(clientIdBytes))
	memberBytesCount += uint64(len(rcIdBytes))
	memberBytesCount += uint64(len(borrowAtBytes))
	memberBytesCount += uint64(len(expireAtBytes))

	vLen := make([]byte, 8)
	binary.BigEndian.PutUint64(vLen[0:8], uint64(memberBytesCount))

	// build
	buf := bytes.NewBuffer(make([]byte, 0, 9+memberBytesCount))
	buf.WriteByte(encoding.VTypeBorrowRecord)
	buf.Write(vLen)
	buf.Write(clientIdBytes)
	buf.Write(rcIdBytes)
	buf.Write(borrowAtBytes)
	buf.Write(expireAtBytes)

	return buf.Bytes(), nil
}

func (rd *BorrowRecord) Decode(b []byte) ([]byte, error) {
	// 校验头部
	if len(b) < 9 {
		return nil, errors.New("Bad encoded format for BorrowRecord, too short")
	}

	// 校验数据类型
	if vType := b[0]; vType != encoding.VTypeBorrowRecord {
		return nil, fmt.Errorf("Bad encoded format for BorrowRecord, VType(%#x) don't match %#x", vType, encoding.VTypeBorrowRecord)
	}

	// 校验数据长度
	vLen := binary.BigEndian.Uint64(b[1:9])
	if uint64(len(b)) < 9+vLen {
		return nil, errors.New("Bad encoded format for BorrowRecord, too short")
	}

	var err error

	// 解析ClientID
	b = b[9:]
	rd.ClientID = encoding.NewBytes(nil)
	b, err = rd.ClientID.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode BorrowRecord.ClientID failed, %v", err)
	}

	// 解析RCID
	rd.RCID = encoding.NewString("")
	b, err = rd.RCID.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode BorrowRecord.RCID failed, %v", err)
	}

	// 解析BorrowAt
	rd.BorrowAt = encoding.NewInt64(0)
	b, err = rd.BorrowAt.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode BorrowRecord.BorrowAt failed, %v", err)
	}

	// 解析ExpireAt
	rd.ExpireAt = encoding.NewInt64(0)
	b, err = rd.ExpireAt.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("Decode BorrowRecord.ExpireAt failed, %v", err)
	}

	return b, nil
}
