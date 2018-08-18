package limiter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// New new one default limiter
func New() Limiter {
	return &limiterV1{
		mutex: &sync.RWMutex{},
		meta:  make(map[string]LimiterMeta),
		stopC: make(chan struct{}),
	}
}

// Limiter is an abstract of rate limiter
type Limiter interface {
	//
	Open() error

	//
	Do(r *Request) Response

	//
	raftlib.FSM
}

// limiterV1 is an implement of Limiter interface
// It use raft to ensure consistency among all nodes
type limiterV1 struct {
	// 读写锁保护meta
	mutex *sync.RWMutex

	// 按照资源类型进行分类
	meta map[string]LimiterMeta // ResourceID ==> LimiterMeta

	//
	raft *raftlib.Raft

	//
	stopC chan struct{}
}

func (l *limiterV1) Open() error {
	// ---------------- Raft集群初始化 ----------------- //
	if g.EnableRaft {
		// 加载集群配置
		confJson := filepath.Join(g.ConfDir, g.RaftClusterConfFile)
		clusterConf, err := raftlib.ReadConfigJSON(confJson)
		if err != nil {
			return fmt.Errorf("Load raft servers' config failed, %v", err)
		}

		raftConf := raftlib.DefaultConfig()
		raftConf.LocalID = raftlib.ServerID(g.LocalID)

		// 创建Raft网络传输
		addr, err := net.ResolveTCPAddr("tcp", g.RaftBind)
		if err != nil {
			return err
		}
		transport, err := raftlib.NewTCPTransport(
			g.RaftBind,       // bindAddr
			addr,             // advertise
			g.RaftTCPMaxPool, // maxPool
			g.RaftTimeout,    // timeout
			os.Stderr,        // logOutput
		)
		if err != nil {
			return fmt.Errorf("Create tcp transport failed, %v", err)
		}

		// 创建持久化Log Entry存储引擎
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(g.RaftDir, "raft.db"))
		if err != nil {
			return fmt.Errorf("Create log store failed, %v", err)
		}

		// 创建持久化快照存储引擎
		snapshotStore, err := raftlib.NewFileSnapshotStore(g.RaftDir, 3, os.Stderr)
		if err != nil {
			return fmt.Errorf("Create file snapshot store failed, %v", err)
		}

		// 创建Raft实例
		raft, err := raftlib.NewRaft(
			raftConf,
			l,
			boltDB,
			boltDB,
			snapshotStore,
			transport,
		)
		if err != nil {
			return fmt.Errorf("Create raft instance failed, %v", err)
		}

		// 启动集群
		future := raft.BootstrapCluster(clusterConf)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Bootstrap raft cluster failed, %v", err)
		}

		// 变量初始化
		l.raft = raft
	} else {
		glog.Info("Raft is disabled")
	}

	// ---------------------- 定时回收&重用资源 ---------------------- //
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-l.stopC:
				return
			case <-ticker.C:
				l.recycle()
			}
		}
	}()

	return nil
}

func (l *limiterV1) IsLeader() bool {
	if g.EnableRaft == false {
		return true
	}
	return l.raft.State() == raftlib.Leader
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (l *limiterV1) Apply(log *raftlib.Log) interface{} {
	switch log.Type {
	case raftlib.LogCommand:
		{
			// 解析远端传过来的命令日志
			var r Request
			if err := json.Unmarshal(log.Data, &r); err != nil {
				return fmt.Errorf("Bad request format for raft command, %v", err)
			}

			// 尝试执行命令
			if rp := l.Do(&r); rp.Err != nil {
				return fmt.Errorf("Try apply command failed, %v", rp.Err)
			}
		}
	default:
		return fmt.Errorf("Unknown LogType(%v)", log.Type)
	}
	return nil
}

// Snapshot 获取对FSM进行快照操作的实例
func (l *limiterV1) Snapshot() (raftlib.FSMSnapshot, error) {
	// TODO:
	return NewLimiterSnapshot(nil), nil
}

// Restore 从快照中恢复LimiterFSM
func (l *limiterV1) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	// 从文件读取快照内容
	buf := bytes.NewBuffer(nil)
	_, err := buf.ReadFrom(rc)
	if err != nil {
		return err
	}

	// Load到对象中
	//var meta map[string]LimiterMeta
	// TODO:
	return nil
}

// Do 执行请求
func (l *limiterV1) Do(r *Request) (rp Response) {
	switch r.Action {
	case ActionRegistQuota:
		rp.Err = l.doRegistQuota(r)
	case ActionBorrow:
		rp.Result, rp.Err = l.doBorrow(r)
	case ActionReturn:
		rp.Err = l.doReturn(r)
	case ActionDead:
		rp.Err = l.doReturnAll(r)
	default:
		rp.Err = errors.New("Action handler not found")
	}
	return rp
}

// doRegistQuota 注册资源配额
func (l *limiterV1) doRegistQuota(r *Request) error {
	// check required
	if len(r.TID) <= 0 {
		return errors.New("Missing `tId` field value")
	}
	if r.Quota <= 0 {
		return errors.New("Invalid `quota` value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try regist
	l.mutex.Lock()
	defer l.mutex.Unlock()

	tIdHex := r.TID.String()
	if _, existed := l.meta[tIdHex]; existed {
		return ErrExisted
	} else {
		l.meta[tIdHex] = NewLimiterMeta(r.TID, r.Quota)
	}

	// commit changes
	if g.EnableRaft {
		cmd, err := json.Marshal(r)
		if err != nil {
			return err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Raft apply error, %v", err)
		}
		if rp := future.Response(); rp != nil {
			return fmt.Errorf("Raft apply response error, %v", rp)
		}
	}

	glog.V(1).Infof("Regist quota for rcType '%s' OK, total:%d", r.TID.String(), r.Quota)

	return nil
}

// doBorrow 借一个资源返回
func (l *limiterV1) doBorrow(r *Request) (string, error) {
	// check required
	if len(r.CID) <= 0 {
		return "", errors.New("Missing `cId` field value")
	}
	if r.Expire <= 0 {
		return "", errors.New("Missing `expire` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return "", ErrNotLeader
	}

	// try borrow
	tIdHex := r.TID.String()
	m, ok := l.meta[tIdHex]
	if !ok {
		return "", fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	rcId, err := m.Borrow(r.CID, r.Expire)
	if err != nil {
		return "", err
	}

	// commit changes
	if g.EnableRaft {
		cmd, err := json.Marshal(r)
		if err != nil {
			return "", err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return "", fmt.Errorf("Raft apply error, %v", err)
		}
		if rp := future.Response(); rp != nil {
			return "", fmt.Errorf("Raft apply response error, %v", rp)
		}
	}

	glog.V(1).Infof("Client[%s] borrow '%s' for %d seconds OK", r.CID.String(), rcId, r.Expire)

	return string(rcId), nil
}

func (l *limiterV1) doReturn(r *Request) error {
	// check required
	if len(r.RCID) <= 0 {
		return errors.New("Missing `rcId` field value")
	}
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try borrow
	tIdHex := r.TID.String()
	m, ok := l.meta[tIdHex]
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	if err := m.Return(r.CID, r.RCID); err != nil {
		return err
	}

	// commit changes
	if g.EnableRaft {
		cmd, err := json.Marshal(r)
		if err != nil {
			return err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Raft apply error, %v", err)
		}
		if rp := future.Response(); rp != nil {
			return fmt.Errorf("Raft apply response error, %v", rp)
		}
	}

	glog.V(1).Infof("Client[%s] return '%s' OK", r.CID.String(), r.RCID, r.Expire)

	return nil
}

func (l *limiterV1) doReturnAll(r *Request) error {
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try borrow
	tIdHex := r.TID.String()
	m, ok := l.meta[tIdHex]
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	if err := m.ReturnAll(r.CID); err != nil {
		return err
	}

	// commit changes
	if g.EnableRaft {
		cmd, err := json.Marshal(r)
		if err != nil {
			return err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Raft apply error, %v", err)
		}
		if rp := future.Response(); rp != nil {
			return fmt.Errorf("Raft apply response error, %v", rp)
		}
	}

	return nil
}

func (l *limiterV1) recycle() {
	// map内容没有被修改，所以这里是读锁
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for _, m := range l.meta {
		m.Recycle()
	}
}
