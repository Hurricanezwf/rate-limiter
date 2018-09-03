package cluster

import "github.com/Hurricanezwf/rate-limiter/meta"

type Interface interface {
	// 集群存储的元数据
	meta.Interface

	// Raft's FSM
	raftlib.FSM

	// IsLeader check if current node is raft leader
	IsLeader() bool

	// LeaderHTTPAddr return current leader's http listen
	LeaderHTTPAddr() string
}
