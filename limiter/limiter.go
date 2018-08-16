package limiter

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	raftlib "github.com/hashicorp/raft"
)

func Default() Limiter {
	l := limiterV1{
		raftApplyTimeout: 30 * time.Second,
	}
	l.handlers = map[ActionType]actionHandler{
		ActionBorrow: l.tryBorrow,
		ActionReturn: l.tryReturn,
		ActionDead:   l.tryReturnAll,
	}
	return &l
}

type Limiter interface {
	//
	Open() error

	//
	Try(r *Request) error

	//
	raftlib.FSM
}

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
	// TODO
	return nil
}

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
