package limiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type actionHandler func(r *Request) error

type limiterV1 struct {
	//
	meta LimiterMeta

	//
	raft *raftlib.Raft

	//
	raftApplyTimeout time.Duration

	//
	handlers map[ActionType]actionHandler
}

func (l *limiterV1) Open() error {
	// 加载集群配置
	confJson := filepath.Join(g.ConfDir, g.RaftClusterConfFile)
	clusterConf, err := raftlib.ReadConfigJSON(confJson)
	if err != nil {
		return fmt.Errorf("Load raft servers' config failed, %v", err)
	}

	// 创建Raft网络传输
	addr, err := net.ResolveTCPAddr("tcp", g.RaftBind)
	if err != nil {
		return err
	}
	transport, err := raftlib.NewTCPTransport(
		g.RaftBind,     // bindAddr
		addr,           // advertise
		3,              // maxPool
		10*time.Second, // timeout
		os.Stderr,      // logOutput
	)
	if err != nil {
		return fmt.Errorf("Create tcp transport failed, %v", err)
	}

	// 创建持久化Log Entry存储引擎
	boltDB, err := raftboltdb.NewBoltStore(g.RaftDir)
	if err != nil {
		return fmt.Errorf("Create log store failed, %v", err)
	}

	// 创建持久化快照存储引擎
	snapshotStore, err := raftlib.NewFileSnapshotStore(g.RaftDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf("Create file snapshot store failed, %v", err)
	}

	// 创建Raft实例
	l.raft, err = raftlib.NewRaft(
		raftlib.DefaultConfig(),
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
	future := l.raft.BootstrapCluster(clusterConf)
	if err = future.Error(); err != nil {
		return fmt.Errorf("Bootstrap raft cluster failed, %v", err)
	}
	return nil
}

// Try 尝试执行action逻辑
func (l *limiterV1) Try(r *Request) error {
	h := l.handlers[r.Action]
	if h == nil {
		return errors.New("Action handler not found")
	}
	return h(r)
}

func (l *limiterV1) IsLeader() bool {
	return l.raft.State() == raftlib.Leader
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (l *limiterV1) Apply(log *raftlib.Log) interface{} {
	// TODO
	return nil
}

// Snapshot 获取对FSM进行快照操作的实例
func (l *limiterV1) Snapshot() (raftlib.FSMSnapshot, error) {
	// TODO:
	return nil, nil
}

// Restore 从快照中恢复LimiterFSM
func (l *limiterV1) Restore(io.ReadCloser) error {
	// TODO:
	return nil
}

func (l *limiterV1) tryBorrow(r *Request) error {
	// check required
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}
	if len(r.PID) <= 0 {
		return errors.New("Missing `pId` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
	}

	// try borrow
	if err := l.meta.Borrow(r.CID, r.PID); err != nil {
		return err
	}

	// commit changes
	cmd, err := json.Marshal(r)
	if err != nil {
		return err
	}

	future := l.raft.Apply(cmd, l.raftApplyTimeout)
	if err = future.Error(); err != nil {
		return fmt.Errorf("Raft apply error, %v", err)
	}
	if rp := future.Response(); rp != nil {
		return fmt.Errorf("Raft apply response error, %v", rp)
	}

	return nil
}

func (l *limiterV1) tryReturn(r *Request) error {
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}
	if len(r.PID) <= 0 {
		return errors.New("Missing `pId` field value")
	}
	return l.meta.Return(r.CID, r.PID)
}

func (l *limiterV1) tryReturnAll(r *Request) error {
	if len(r.CID) <= 0 {
		return errors.New("Missing `cId` field value")
	}
	if len(r.PID) <= 0 {
		return errors.New("Missing `pId` field value")
	}
	return l.meta.ReturnAll(r.CID)
}
