package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

// NewQueue 构造一个非并发安全的队列实例
func NewQueue() *PB_Queue {
	q := &PB_Queue{}
	return q.Init()
}

// Init 清空重置链表
func (q *PB_Queue) Init() *PB_Queue {
	q.Reset()
	return q
}

// PushBack 向队列中添加一个值
func (q *PB_Queue) PushBack(value proto.Message) (*PB_Element, error) {
	v, err := ptypes.MarshalAny(value)
	if err != nil {
		return nil, err
	}
	return q.pushBackAny(v), nil
}

// pushBackAny 添加一个类型是*any.Any的元素到队列
func (q *PB_Queue) pushBackAny(v *any.Any) *PB_Element {
	element := &PB_Element{
		Next:  nil,
		Prev:  q.Tail,
		Value: v,
	}
	if q.Size == 0 {
		q.Head = element
		q.Tail = element
	} else {
		q.Tail.Next = element
	}

	q.Size++

	return element
}

// PushBackList 向队列中添加另一个队列的元素拷贝
func (q *PB_Queue) PushBackList(other *PB_Queue) {
	if other == nil {
		return
	}
	for itr := other.Front(); itr != nil; itr = itr.Next {
		q.pushBackAny(itr.Value)
	}
}

// PopFront 弹出队首元素，如果队列为空，则返回nil
func (q *PB_Queue) PopFront() *PB_Element {
	e := q.Front()
	if e != nil {
		q.Remove(e)
	}
	return e
}

// Front 获取链表第一个元素，如果为空链表则返回nil
func (q *PB_Queue) Front() *PB_Element {
	return q.Head
}

// Back 获取链表的尾部元素，如果链表为空，则返回空
func (q *PB_Queue) Back() *PB_Element {
	return q.Tail
}

// Remove 从链表中删除一个非空元素
func (q *PB_Queue) Remove(e *PB_Element) {
	if e == nil {
		panic("Element to remove is nil")
	}

	for itr := q.Front(); itr != nil; itr = itr.Next {
		if itr != e {
			continue
		}

		q.Size--

		// 删除了唯一的结点
		if q.Head == q.Tail {
			q.Head, q.Tail = nil, nil
			return
		}

		// 删除了back结点
		if itr.Next == nil {
			itr.Prev.Next = nil
			q.Tail = itr.Prev
			itr.Prev = nil
			return
		}

		// 删除front结点
		if itr.Prev == nil {
			q.Head = itr.Next
			itr.Next = nil
			return
		}
	}
}

// Len 获取链表长度
func (q *PB_Queue) Len() uint32 {
	return q.Size
}
