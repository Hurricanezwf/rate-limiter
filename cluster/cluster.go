package cluster

import "github.com/Hurricanezwf/rate-limiter/meta"

type Interface interface {
	// Open 启用集群
	Open() error

	// Close 关闭集群
	Close() error

	// 集群存储的元数据
	meta.Interface

	// Raft's FSM
	raftlib.FSM

	// IsLeader check if current node is raft leader
	IsLeader() bool

	// LeaderHTTPAddr return current leader's http listen
	LeaderHTTPAddr() string
}
