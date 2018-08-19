package limiter

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"

	. "github.com/Hurricanezwf/rate-limiter/proto"
)

type Serializer interface {
	// Encode encode object to []byte
	Encode() ([]byte, error)

	// Decode decode from []byte to object
	Decode([]byte) ([]byte, error)
}

type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{
		l: list.New(),
	}
}

func (q *Queue) Init() {
	q.l.Init()
}

// PushBack push id to the back of the queue
func (q *Queue) PushBack(id ResourceID) {
	q.l.PushBack(id)
}

// PopFront if queue is empty, false will be returnded
func (q *Queue) PopFront() (ResourceID, bool) {
	if q.l.Len() <= 0 {
		return ResourceID(""), false
	}
	e := q.l.Front()
	q.l.Remove(e)
	return e.Value.(ResourceID), true
}

// Encode encode queue to []byte
//
// > [4 bytes] queue len
// > [1 byte]  element_1's len
// > [N bytes] element_1's content
// > ...
func (q *Queue) Encode() ([]byte, error) {
	queueLen := make([]byte, 4)
	binary.BigEndian.PutUint32(queueLen, uint32(q.l.Len()))

	buf := bytes.NewBuffer(nil)
	buf.Write(queueLen)

	for e := q.l.Front(); e != nil; e = e.Next() {
		rcId := e.Value.(ResourceID)
		rcIdBytes, err := rcId.Encode()
		if err != nil {
			return nil, fmt.Errorf("Encode ResourceID(%s) to bytes failed, %v", rcId, err)
		}
		buf.Write(rcIdBytes)
	}
	return buf.Bytes(), nil
}

func (q *Queue) Decode(b []byte) ([]byte, error) {
	if len(b) < 4 {
		return b, errors.New("Bad Queue byte sequence")
	}

	q.Init()

	queueLen := binary.BigEndian.Uint32(b[0:5])
	queueBytes := b[5 : 5+queueLen]
	remainingBytes := b[5+queueLen:]

	for i := uint32(0); i < queueLen; i++ {
		if len(queueBytes) <= 0 {
			return b, fmt.Errorf("Missing queue data in byte sequence, %d should have, missing %d elements", queueLen, queueLen-i)
		}

		var err error
		var rcId ResourceID

		queueBytes, err = rcId.Decode(queueBytes)
		if err != nil {
			return b, err
		}
		q.PushBack(rcId)
	}
	return remainingBytes, nil
}
