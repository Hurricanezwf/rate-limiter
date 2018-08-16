package limiter

import raftlib "github.com/hashicorp/raft"

type LimiterSnapshot struct {
	data []byte
}

func NewLimiterSnapshot(data []byte) *LimiterSnapshot {
	return &LimiterSnapshot{
		data: data,
	}
}

func (s *LimiterSnapshot) Persist(sink raftlib.SnapshotSink) error {
	// TODO:
	return nil
}

func (s *LimiterSnapshot) Relase() {
	// TODO
}
