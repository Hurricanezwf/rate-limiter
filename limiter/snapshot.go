package limiter

import (
	"io"

	raftlib "github.com/hashicorp/raft"
)

type LimiterSnapshot struct {
	dataToPersist io.Reader
}

func NewLimiterSnapshot(dataToPersist io.Reader) *LimiterSnapshot {
	return &LimiterSnapshot{
		dataToPersist: dataToPersist,
	}
}

func (s *LimiterSnapshot) Persist(sink raftlib.SnapshotSink) error {
	_, err := io.Copy(sink, s.dataToPersist)
	if err != nil {
		sink.Cancel()
	} else {
		sink.Close()
	}
	return err
}

func (s *LimiterSnapshot) Release() {
	// do nothing
}
