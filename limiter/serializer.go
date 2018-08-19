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
	// 元素类型
	vType byte

	*list.List
}

func NewQueue(vType byte) *Queue {
	return &Queue{
		List:  list.New(),
		vType: vType,
	}
}

// PushBack push id to the back of the queue
func (q *Queue) PushBack(v interface{}) {
	if ValueTypeMatch(q.vType, v) == false {
		panic(fmt.Sprintf("value's type is not %v", q.vType))
	}
	q.List.PushBack(v)
}

// PopFront if queue is empty, false will be returnded
func (q *Queue) PopFront() (interface{}, bool) {
	if q.List.Len() <= 0 {
		return nil, false
	}
	e := q.List.Front()
	q.List.Remove(e)
	return e.Value, true
}

func (q *Queue) PushBackQueue(toPush *Queue) {
	if q.vType != toPush.vType {
		panic("ValueType not equal")
	}
	q.PushBackList(toPush.List)
}

// Encode encode queue to []byte
//
// > [1 byte]  queue element type
// > [4 bytes] queue len
// > [1 byte]  element_1's len
// > [N bytes] element_1's content
// > ...
func (q *Queue) Encode() ([]byte, error) {
	header := make([]byte, 5)
	header[0] = byte(q.vType)
	binary.BigEndian.PutUint32(header[1:5], uint32(q.List.Len()))

	buf := bytes.NewBuffer(nil)
	buf.Write(header)

	for e := q.List.Front(); e != nil; e = e.Next() {
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
	if len(b) < 5 {
		return b, errors.New("Bad Queue byte sequence")
	}

	vType := b[0]
	if vType != q.vType {
		panic(fmt.Sprintf("ValueType(%v) != %v", vType, q.vType))
	}

	q.Init()

	queueLen := binary.BigEndian.Uint32(b[1:6])
	queueBytes := b[6 : 6+queueLen]
	remainingBytes := b[6+queueLen:]

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
