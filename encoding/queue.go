package encoding

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"errors"
	"fmt"
)

// Queue implement of Serializer interface{}
type Queue struct {
	l *list.List
}

func NewQueue() *Queue {
	return &Queue{
		l: list.New(),
	}
}

// Init initializes or clears queue q.
func (q *Queue) Init() {
	q.l.Init()
}

// Encode encode queue to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [4 bytes] element's count of the queue
// > --------------------------------------------
// > [4 bytes] byte count of the first element's encoded result
// > [N bytes] the first element's encoded result
// > --------------------------------------------
// > ...
// > ...
// > ...
func (q *Queue) Encode() ([]byte, error) {
	header := make([]byte, 5)

	// 数据类型
	header[0] = VTypeQueue

	// 队列长度
	binary.BigEndian.PutUint32(header[1:5], uint32(q.l.Len()))

	// 写入队列数据
	buf := bytes.NewBuffer(header)
	buf.Grow(10240)

	for itr := q.l.Front(); itr != nil; itr = itr.Next() {
		s := itr.Value.(Serializer)
		b, err := s.Encode()
		if err != nil {
			return nil, fmt.Errorf("Encode queue's data failed, %v", err)
		}
		buf.Write(b)
	}

	return buf.Bytes(), nil
}

func (q *Queue) Decode(b []byte) ([]byte, error) {
	if len(b) < 5 {
		return nil, errors.New("Bad encoded format for queue, too short")
	}

	// 数据类型
	if vType := b[0]; vType != VTypeQueue {
		return nil, fmt.Errorf("Bad encoded format for queue, VType(%x) don't match %x", vType, VTypeQueue)
	}

	// 队列长度
	qLen := binary.BigEndian.Uint32(b[1:5])

	// 解析队列数据
	q.Init()
	b = b[5:]
	for i := uint32(0); i < qLen; i++ {
		if len(b) < 1 {
			return nil, errors.New("Bad encoded format for queue's element, too short")
		}

		// 解析数据类型并获取serializer
		elementType := b[0]
		element, err := SerializerFactory(elementType)
		if err != nil {
			return nil, fmt.Errorf("Get serializer failed, %v", err)
		}

		// 解析队列元素
		b, err = element.Decode(b)
		if err != nil {
			return nil, fmt.Errorf("Decode queue's element failed, %v", err)
		}

		// 加入队列
		q.PushBack(element)
	}
	return b, nil
}

func (q *Queue) Len() int {
	return q.l.Len()
}

func (q *Queue) Front() QueueElement {
	return QueueElement{
		element: q.l.Front(),
	}
}

func (q *Queue) PushBack(v Serializer) {
	q.l.PushBack(v)
}

// PopFront pop the first element from queue, if not found, then false will be returned
func (q *Queue) PopFront() (interface{}, bool) {
	if q.l.Len() <= 0 {
		return nil, false
	}
	e := q.l.Front()
	q.l.Remove(e)
	return e.Value, true
}

func (q *Queue) PushBackQueue(q2 *Queue) {
	q.l.PushBackList(q2.l)
}

func (q *Queue) Remove(e QueueElement) {
	if e.IsNil() == false {
		q.l.Remove(e.element)
	}
}

// QueueElement is an element of a queue
type QueueElement struct {
	element *list.Element
}

func (e QueueElement) IsNil() bool {
	return e.element == nil
}

func (e QueueElement) Next() QueueElement {
	return QueueElement{
		element: e.element.Next(),
	}
}

func (e QueueElement) Value() interface{} {
	return e.element.Value
}
