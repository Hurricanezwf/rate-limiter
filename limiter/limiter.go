package limiter

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/g"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// New new one default limiter
func New() Limiter {
	return &limiterV1{}
}

// Limiter is an abstract of rate limiter
type Limiter interface {
	//
	Open() error

	//
	Do(r *Request) error

	//
	raftlib.FSM
}

// limiterV1 is an implement of Limiter interface
// It use raft to ensure consistency among all nodes
type limiterV1 struct {
	// 读写锁保护meta
	mutex sync.RWMutex

	// 按照资源类型进行分类
	meta map[string]LimiterMeta // ResourceID ==> LimiterMeta

	//
	raft *raftlib.Raft
}

func (l *limiterV1) Open() error {
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
	l.mutex.Lock()
	l.raft = raft
	l.meta = make(map[string]LimiterMeta)
	l.mutex.Unlock()
	return nil
}

func (l *limiterV1) IsLeader() bool {
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
			if err := l.Do(&r); err != nil {
				return fmt.Errorf("Try apply command failed, %v", err)
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
func (l *limiterV1) Do(r *Request) error {
	switch r.Action {
	case ActionRegistQuota:
		return l.doRegistQuota(r)
	case ActionBorrow:
		return l.doBorrow(r)
	case ActionReturn:
		return l.doReturn(r)
	case ActionDead:
		return l.doReturnAll(r)
	}
	return errors.New("Action handler not found")
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

	tIdHex := hex.EncodeToString(r.TID)
	if _, existed := l.meta[tIdHex]; existed {
		return ErrExisted
	} else {
		l.meta[tIdHex] = NewLimiterMeta(r.TID, r.Quota)
	}

	return nil
}

func (l *limiterV1) doBorrow(r *Request) error {
	// check required
	if len(r.TID) <= 0 {
		return errors.New("Missing `tId` field value")
	}
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}
	if r.Expire <= 0 {
		return errors.New("Missing `expire` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try borrow
	tIdHex := hex.EncodeToString(r.TID)
	m, ok := l.meta[tIdHex]
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	if _, err := m.Borrow(r.TID, r.CID, r.Expire); err != nil {
		return err
	}

	// commit changes
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

	return nil
}

func (l *limiterV1) doReturn(r *Request) error {
	// check required
	if len(r.TID) <= 0 {
		return errors.New("Missing `tId` field value")
	}
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try borrow
	tIdHex := hex.EncodeToString(r.TID)
	m, ok := l.meta[tIdHex]
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	if err := m.Return(r.TID, r.CID); err != nil {
		return err
	}

	// commit changes
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
	tIdHex := hex.EncodeToString(r.TID)
	m, ok := l.meta[tIdHex]
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	}
	if err := m.ReturnAll(r.CID); err != nil {
		return err
	}

	// commit changes
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

	return nil
}
