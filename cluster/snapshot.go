package cluster

import (
	"io"

	raftlib "github.com/hashicorp/raft"
)

type Snapshot struct {
	dataToPersist io.Reader
}

func NewSnapshot(dataToPersist io.Reader) *Snapshot {
	return &Snapshot{
		dataToPersist: dataToPersist,
	}
}

func (s *Snapshot) Persist(sink raftlib.SnapshotSink) error {
	_, err := io.Copy(sink, s.dataToPersist)
	if err != nil {
		sink.Cancel()
	} else {
		sink.Close()
	}
	return err
}

func (s *Snapshot) Release() {
	// do nothing
}
