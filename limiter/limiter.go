package limiter

import (
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	raftlib "github.com/hashicorp/raft"
)

type Limiter interface {
	//
	Open() error

	//
	Try(r *Request) error

	//
	raftlib.FSM
}

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
