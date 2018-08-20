package encoding

import "container/list"

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

func (q *Queue) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (q *Queue) Decode(b []byte) ([]byte, error) {
	// TODO
	return nil, nil
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
