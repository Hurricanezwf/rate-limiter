package meta

import (
	"cmp/public-cloud/proxy-layer/logging/glog"
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/Hurricanezwf/rate-limiter/proto"
)

// rcManager 存放了一种类型资源的元数据
type rcManager struct {
	// 并发安全锁
	gLock *sync.RWMutex

	// 资源类型
	rcType []byte

	// 资源配额
	quota uint32

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

func newRCManager(rcType []byte, quota uint32) *rcManager {
	return &rcManager{
		gLock:     &sync.RWMutex{},
		rcType:    rcType,
		quota:     quota,
		canBorrow: list.New(),
		recycled:  list.New(),
		used:      make(map[string]make(map[string]*borrowRecord)),
		usedCount: 0,
	}
}

func (mgr *rcManager) safeBorrow(rcType, clientId []byte, expire int64) (string, error) {
	clientIdHex := encoding.BytesToStringHex(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 验证配额是否足够
	e := mgr.canBorrow.Front()
	if e != nil {
		return "", proto.ErrQuotaNotEnough
	}

	// 构造出借记录
	rcId := e.Value.(string)
	nowTs := time.Now().Unix()
	rd := &borrowRecord{
		borrowAt: nowTs,
		expireAt: nowTs + expire,
	}

	// 执行借出
	rdTable := mgr.used[clientIdHex]
	if rdTable == nil {
		rdTable = make(map[string]*borrowRecord)
	}
	rdTable[rcId] = rd
	mgr.canBorrow.Remove(e)
	mgr.usedCount++

	return rcId, nil
}

func (mgr *rcManager) safeReturn(clientId []byte, rcId string) error {
	clientIdHex := encoding.BytesToStringHex(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 安全性检测
	if mgr.recycled.Len() >= mgr.quota {
		return fmt.Errorf("There's something wrong, recycled queue len(%d) >= quota(%d)", mgr.recycled.Len(), mgr.quota)
	}

	// 归还资源
	// 从出借记录中删除并加入recycle队列
	rdTable, exist := mgr.used[clientIdHex]
	if !exist {
		return proto.ErrNoBorrowRecordFound
	}
	if _, exist = rdTable[rcId]; !exist {
		return proto.ErrNoBorrowRecordFound
	}
	delete(rdTable, rcId)
	mgr.usedCount--
	if err := m.safePutRecycle(rcId); err != nil {
		glog.Fatalf("%v when merge expired record into recycled queue", err)
	}

	if len(rdTable) <= 0 {
		// 没有该用户的借入记录时，清除防止累积增长
		delete(mgr.used, clientIdHex)
	}

	return nil
}

func (mgr *rcManager) safeReturnAll(clientId []byte) (uint32, error) {
	rcTypeHex := encoding.BytesToStringHex(rcType)
	clientIdHex := encoding.BytesToStringHex(clientId)

	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	rdTable, _ := mgr.used[clientIdHex]
	count := len(rdTable)
	if count > 0 {
		for rcId, _ := range rdTable {
			if err := m.safePutRecycle(rcId); err != nil {
				glog.Fatalf("%v when merge expired record into recycled queue", err)
			}
		}
		m.usedCount -= count
		delete(mgr.used, clientIdHex)
	}

	return count, nil
}

func (mgr *rcManager) recycle() {
	mgr.gLock.Lock()
	defer mgr.gLock.Unlock()

	// 过期资源清理
	nowTs := time.Now().Unix()
	for clientIdHex, rdTable := range mgr.used {
		for rcId, rd := range rdTable {
			if rd.expireAt > nowTs {
				continue
			}
			glog.V(2).Infof("'%s' borrowed by client '%s' is expired, force to recycle", rcId, clientIdHex)
			delete(rdTable, rcId)
			mgr.usedCount--
			if err := mgr.safePutRecycle(rcId); err != nil {
				glog.Fatalf("%v when merge expired record into recycled queue", err)
			}
		}
		if len(rdTable) <= 0 {
			delete(mgr.used, clientIdHex)
		}
	}

	// 资源重用
	if count := mgr.recycled.Len(); count > 0 {
		if err := mgr.safePutCanBorrowWith(mgr.recycled); err != nil {
			glog.Fatalf("%v when merge recycled into canBorrow", err)
		}
		mgr.recycled.Init()
		glog.V(2).Infof("Refresh %d resources to canBorrow queue because of client's return", count)
	}
}

func (mgr *rcManager) safePutCanBorrowWith(recycled *list.List) error {
	if mgr.canBorrow.Len()+recycled.Len() < mgr.quota {
		mgr.canBorrow.PushBack(recycled)
		return nil
	}
	return proto.ErrQuotaOverflow
}

func (mgr *rcManager) safePutRecycle(rcId string) error {
	if mgr.canBorrow.Len()+mgr.recycled.Len() < mgr.quota {
		mgr.recycled.PushBack(rcId)
		return nil
	}
	return proto.ErrQuotaOverflow
}
