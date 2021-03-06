package meta

import (
	"container/list"
	"errors"
	"fmt"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	pb "github.com/golang/protobuf/proto"
)

// metaV2 是Interface接口的具体实现
type metaV2 struct {
	// 全局锁
	gLock *sync.RWMutex

	// 资源元数据容器
	// rcType ==> *rcManager
	mgr map[string]*rcManager
}

func newMetaV2() Interface {
	return &metaV2{
		gLock: &sync.RWMutex{},
		mgr:   make(map[string]*rcManager),
	}
}

func (m *metaV2) safeFindManager(rcTypeStr string) *rcManager {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	return m.mgr[rcTypeStr]
}

func (m *metaV2) RegistQuota(rcType []byte, quota uint32, resetInterval, timestamp int64) error {
	m.gLock.Lock()
	defer m.gLock.Unlock()

	// 验证是否已经注册
	rcTypeStr := encoding.BytesToString(rcType)
	rcMgr := m.mgr[rcTypeStr]
	if rcMgr != nil {
		return ErrExisted
	}

	// 初始化资源元数据
	rcMgr = newRCManager(rcType, quota, resetInterval)
	for i := uint32(0); i < quota; i++ {
		rcMgr.canBorrow.PushBack(MakeResourceID(rcType, i))
	}

	m.mgr[rcTypeStr] = rcMgr

	return nil
}

func (m *metaV2) DeleteQuota(rcType []byte) error {
	m.gLock.Lock()
	defer m.gLock.Unlock()

	rcTypeStr := encoding.BytesToString(rcType)
	rcMgr := m.mgr[rcTypeStr]
	if rcMgr == nil {
		return ErrResourceNotRegisted
	}
	delete(m.mgr, rcTypeStr)
	return nil
}

func (m *metaV2) Borrow(rcType, clientId []byte, expire, timestamp int64) (string, error) {
	rcTypeStr := encoding.BytesToString(rcType)
	rcMgr := m.safeFindManager(rcTypeStr)
	if rcMgr == nil {
		return "", ErrResourceNotRegisted
	}
	return rcMgr.safeBorrow(rcType, clientId, expire, timestamp)
}

func (m *metaV2) Return(clientId []byte, rcId string) error {
	rcType, err := ResolveResourceID(rcId)
	if err != nil {
		return fmt.Errorf("Resolve resource id failed, %v", err)
	}

	rcTypeStr := encoding.BytesToString(rcType)
	rcMgr := m.safeFindManager(rcTypeStr)
	if rcMgr == nil {
		return ErrResourceNotRegisted
	}
	return rcMgr.safeReturn(clientId, rcId)
}

func (m *metaV2) ReturnAll(rcType, clientId []byte) (uint32, error) {
	rcTypeStr := encoding.BytesToString(rcType)
	rcMgr := m.safeFindManager(rcTypeStr)
	if rcMgr == nil {
		return 0, ErrResourceNotRegisted
	}
	return rcMgr.safeReturnAll(clientId)
}

func (m *metaV2) Recycle(timestamp int64) {
	m.gLock.RLock()
	defer m.gLock.RUnlock()
	for _, rcMgr := range m.mgr {
		rcMgr.safeRecycle(timestamp)
	}
}

func (m *metaV2) ResourceList(rcType []byte) ([]*ResourceDetail, error) {
	var err error
	var details []*ResourceDetail

	switch len(rcType) {
	case 0:
		// 查询全量资源的详情
		{
			m.gLock.RLock()
			defer m.gLock.RUnlock()

			for _, rcMgr := range m.mgr {
				details = append(details, rcMgr.safeGetDetail())
			}
		}
	default:
		// 查询指定的资源详情
		{
			rcTypeStr := encoding.BytesToString(rcType)
			rcMgr := m.safeFindManager(rcTypeStr)
			if rcMgr == nil {
				return nil, ErrResourceNotRegisted
			}
			details = append(details, rcMgr.safeGetDetail())
		}
	}

	return details, err
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

	for rcTypeStr, rcMgr := range m.mgr {
		pbMgr := &PB_Manager{
			RCType:        rcMgr.rcType,
			Quota:         rcMgr.quota,
			ResetInterval: rcMgr.resetInterval,
			LastReset:     rcMgr.lastReset,
			CanBorrow:     make([]string, 0, rcMgr.canBorrow.Len()),
			Recycled:      make([]string, 0, rcMgr.recycled.Len()),
			Used:          make([]*PB_BorrowRecord, 0, rcMgr.usedCount),
			UsedCount:     rcMgr.usedCount,
		}
		pb.Value[rcTypeStr] = pbMgr

		for itr := rcMgr.canBorrow.Front(); itr != nil; itr = itr.Next() {
			rcId := itr.Value.(string)
			pbMgr.CanBorrow = append(pbMgr.CanBorrow, rcId)
		}
		for itr := rcMgr.recycled.Front(); itr != nil; itr = itr.Next() {
			rcId := itr.Value.(string)
			pbMgr.Recycled = append(pbMgr.Recycled, rcId)
		}
		for clientIdStr, rdTable := range rcMgr.used {
			clientId, err := encoding.StringToBytes(clientIdStr)
			if err != nil {
				glog.Warningf("Convert clientIdStr '%s' to bytes failed, %v", clientIdStr, err)
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

	for rcTypeStr, pbMgr := range pb.Value {
		rcMgr := newRCManager(pbMgr.RCType, pbMgr.Quota, pbMgr.ResetInterval)
		rcMgr.quota = pbMgr.Quota
		rcMgr.resetInterval = pbMgr.ResetInterval
		rcMgr.lastReset = pbMgr.LastReset
		rcMgr.usedCount = pbMgr.UsedCount

		for _, rcId := range pbMgr.CanBorrow {
			rcMgr.canBorrow.PushBack(rcId)
		}
		for _, rcId := range pbMgr.Recycled {
			rcMgr.recycled.PushBack(rcId)
		}
		for _, pbRd := range pbMgr.Used {
			clientIdStr := encoding.BytesToString(pbRd.ClientID)
			rdTable := rcMgr.used[clientIdStr]
			if rdTable == nil {
				rdTable = make(map[string]*borrowRecord)
			}
			rdTable[pbRd.RCID] = &borrowRecord{
				borrowAt: pbRd.BorrowAt,
				expireAt: pbRd.ExpireAt,
			}
			rcMgr.used[clientIdStr] = rdTable
		}
		m.mgr[rcTypeStr] = rcMgr
	}
}

// rcManager 存放了一种类型资源的元数据
type rcManager struct {
	// 并发安全锁
	gLock *sync.RWMutex

	// 资源类型
	rcType []byte
	// 资源配额
	quota uint32
	// 重置资源配额的时间间隔, 单位秒
	resetInterval int64
	// 最近一次重置时间
	lastReset int64

	// 可用资源队列
	canBorrow *list.List

	// 待复用资源队列
	recycled *list.List

	// 已使用资源记录
	// map[clientId]map[rcId]*borrowRecord
	used map[string]map[string]*borrowRecord

	// 该资源已使用总量统计
	usedCount uint32
}

// borrowRecord 资源借出记录
type borrowRecord struct {
	borrowAt int64
	expireAt int64
}

func newRCManager(rcType []byte, quota uint32, resetInterval int64) *rcManager {
	return &rcManager{
		gLock:         &sync.RWMutex{},
		rcType:        rcType,
		quota:         quota,
		resetInterval: resetInterval,
		lastReset:     -1,
		canBorrow:     list.New(),
		recycled:      list.New(),
		used:          make(map[string]map[string]*borrowRecord),
		usedCount:     0,
	}
}

func (mgr *rcManager) safeBorrow(rcType, clientId []byte, expire, timestamp int64) (string, error) {
	clientIdStr := encoding.BytesToString(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 验证配额是否足够
	e := mgr.canBorrow.Front()
	if e == nil {
		return "", ErrQuotaNotEnough
	}

	// 构造出借记录
	rcId := e.Value.(string)
	rd := &borrowRecord{
		borrowAt: timestamp,
		expireAt: timestamp + expire,
	}

	// 执行借出
	rdTable := mgr.used[clientIdStr]
	if rdTable == nil {
		rdTable = make(map[string]*borrowRecord)
	}
	rdTable[rcId] = rd
	mgr.used[clientIdStr] = rdTable
	mgr.canBorrow.Remove(e)
	mgr.usedCount++

	// 从成功借出第一个资源开始进行重置计时
	if mgr.lastReset <= 0 {
		mgr.lastReset = timestamp
	}

	return rcId, nil
}

func (mgr *rcManager) safeReturn(clientId []byte, rcId string) error {
	clientIdStr := encoding.BytesToString(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 安全性检测
	if uint32(mgr.recycled.Len()) >= mgr.quota {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", mgr.recycled.Len(), mgr.quota)
	}

	// 归还资源
	// 从出借记录中删除并加入recycle队列
	rdTable, exist := mgr.used[clientIdStr]
	if !exist {
		return ErrNoBorrowRecordFound
	}
	if _, exist = rdTable[rcId]; !exist {
		return ErrNoBorrowRecordFound
	}
	if err := mgr.safePutRecycle(rcId); err != nil {
		glog.Warningf("%v when merge expired record into recycled queue, canBorrow:%d, recycled:%d, quota:%d", err, mgr.canBorrow.Len(), mgr.recycled.Len(), mgr.quota)
		return err
	}
	delete(rdTable, rcId)
	mgr.usedCount--

	if len(rdTable) <= 0 {
		// 没有该用户的借入记录时，清除防止累积增长
		delete(mgr.used, clientIdStr)
	}

	return nil
}

func (mgr *rcManager) safeReturnAll(clientId []byte) (uint32, error) {
	clientIdStr := encoding.BytesToString(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	rdTable, _ := mgr.used[clientIdStr]
	count := uint32(len(rdTable))
	if count > 0 {
		for rcId, _ := range rdTable {
			if err := mgr.safePutRecycle(rcId); err != nil {
				glog.Warningf("%v when merge expired record into recycled queue, canBorrow:%d, recycled:%d, quota:%d", err, mgr.canBorrow.Len(), mgr.recycled.Len(), mgr.quota)
				continue
			}
		}
		mgr.usedCount -= count
		delete(mgr.used, clientIdStr)
	}

	return count, nil
}

func (mgr *rcManager) safeRecycle(timestamp int64) {
	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 还没有资源借出，此时跳过此步骤
	if mgr.lastReset <= 0 {
		return
	}

	// 过期资源清理
	nowTs := timestamp
	for clientIdStr, rdTable := range mgr.used {
		for rcId, rd := range rdTable {
			if rd.expireAt > nowTs {
				continue
			}
			glog.V(2).Infof("'%s' borrowed by client '%s' is expired, force to recycle", rcId, clientIdStr)
			if err := mgr.safePutRecycle(rcId); err != nil {
				glog.Warningf("%v when merge expired record into recycled queue, canBorrow:%d, recycled:%d, quota:%d, rcType:'%s'",
					err, mgr.canBorrow.Len(), mgr.recycled.Len(), mgr.quota, encoding.BytesToString(mgr.rcType))
				continue
			}
			delete(rdTable, rcId)
			mgr.usedCount--
		}
		if len(rdTable) <= 0 {
			delete(mgr.used, clientIdStr)
		}
	}

	// 资源重用
	if nowTs >= mgr.lastReset+mgr.resetInterval {
		count := mgr.recycled.Len()
		if count <= 0 {
			return
		}
		if err := mgr.safePutCanBorrowWith(mgr.recycled); err != nil {
			glog.Warningf("%v when merge recycled into canBorrow, canBorrow:%d, recycled:%d, quota:%d, rcType:'%s'",
				err, mgr.canBorrow.Len(), mgr.recycled.Len(), mgr.quota, encoding.BytesToString(mgr.rcType))
		}
		mgr.recycled.Init()
		mgr.lastReset = nowTs
		glog.V(2).Infof("Refresh %d resources to canBorrow queue because of client's return, rcType='%s'", count, encoding.BytesToString(mgr.rcType))
	}
}

func (mgr *rcManager) safePutCanBorrowWith(recycled *list.List) error {
	if uint32(mgr.canBorrow.Len()+recycled.Len()) <= mgr.quota {
		mgr.canBorrow.PushBackList(recycled)
		return nil
	}
	return ErrQuotaOverflow
}

func (mgr *rcManager) safePutRecycle(rcId string) error {
	if uint32(mgr.canBorrow.Len()+mgr.recycled.Len()) < mgr.quota {
		mgr.recycled.PushBack(rcId)
		return nil
	}
	return ErrQuotaOverflow
}

func (mgr *rcManager) safeGetDetail() *ResourceDetail {
	mgr.gLock.RLock()
	defer mgr.gLock.RUnlock()

	return &ResourceDetail{
		RCType:         mgr.rcType,
		Quota:          mgr.quota,
		ResetInterval:  mgr.resetInterval,
		CanBorrowCount: uint32(mgr.canBorrow.Len()),
		RecycledCount:  uint32(mgr.recycled.Len()),
		UsedCount:      mgr.usedCount,
	}
}
