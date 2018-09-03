package cluster

import (
	"github.com/Hurricanezwf/rate-limiter/meta"
	raftlib "github.com/hashicorp/raft"
)

func init() {
	RegistBuilder("v2", newClusterV2)
}

type Interface interface {
	// Open 启用集群
	Open() error

	// 集群存储的元数据
	meta.Interface

	// Raft's FSM
	raftlib.FSM

	// IsLeader check if current node is raft leader
	IsLeader() bool

	// LeaderHTTPAddr return current leader's http listen
	LeaderHTTPAddr() string
}

// New 新建一个cluster接口实例
func New(name string) Interface {
	if builders == nil {
		return nil
	}
	if f := builders[name]; f != nil {
		return f()
	}
	return nil
}

// clusterV2 cluster.Interface的具体实现
type clusterV2 struct {
	m meta.Insterface
}
