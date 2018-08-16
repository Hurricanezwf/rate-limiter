package raft

import (
	"io"

	"github.com/Hurricanezwf/rate-limiter/limiter"
	raftlib "github.com/hashicorp/raft"
)

// LimiterFSM 实现了raft库的FSM接口, 主要负责保证集群数据一致性
type LimiterFSM struct {
	L limiter.Limiter
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (fsm *LimiterFSM) Apply(log *raftlib.Log) interface{} {
	// TODO
	return nil
}

// Snapshot 获取对FSM进行快照操作的实例
func (fsm *LimiterFSM) Snapshot() (raftlib.FSMSnapshot, error) {
	// TODO:
	return nil, nil
}

// Restore 从快照中恢复LimiterFSM
func (fsm *LimiterFSM) Restore(io.ReadCloser) error {
	// TODO:
	return nil
}
