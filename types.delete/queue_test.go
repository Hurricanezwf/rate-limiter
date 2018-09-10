package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestQueuePushBack(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	for itr := q.Front(); itr != nil; itr = itr.Next {
		t.Logf("%s\n", getString(itr))
	}
}

func TestQueuePushBackWithDebug(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	q.debug()
}

func TestQueueFrontAndBack(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	t.Logf("Front: %s\n", getString(q.Front()))
	t.Logf("Back: %s\n", getString(q.Back()))
}

func TestQueueInit(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	t.Logf("Before Init: Ptr=%p, Len=%d\n", q, q.Len())
	q.Init()
	t.Logf("After  Init: Ptr=%p, Len=%d\n", q, q.Len())
}

func TestQueuePopFront(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	for e := q.PopFront(); e != nil; e = q.PopFront() {
		t.Logf("%s\n", getString(e))
	}
}

func TestQueuePushBackList(t *testing.T) {
	q1 := NewQueue()
	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q1.PushBack(str)
	}

	q2 := NewQueue()
	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q2.PushBack(str)
	}

	q1.PushBackList(q2)

	for e := q1.PopFront(); e != nil; e = q1.PopFront() {
		t.Logf("%s\n", getString(e))
	}
}

func TestQueuePushFront(t *testing.T) {
	q := NewQueue()
	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushFront(str)
	}

	for itr := q.Head; itr != nil; itr = itr.Next {
		t.Logf("%s\n", getString(itr))
	}
}

func TestQueueRemove(t *testing.T) {
	q := NewQueue()
	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushFront(str)
	}

	for tail := q.Tail; tail != nil; {
		prev := tail.Prev
		q.Remove(tail)
		tail = prev
		t.Logf("Len:%d\n", q.Len())
	}
}

func TestQueueMarshal(t *testing.T) {
	q := NewQueue()
	q.PushBack(NewString("rc#0"))
	q.PushBack(NewString("rc#1"))

	b, err := json.Marshal(q)
	//b, err := proto.Marshal(q)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("Len:%d, Bytes: %#v", len(b), b)
}

func getString(e *PB_Element) string {
	str := NewString("")
	UnmarshalAny(e.Value, str)
	return str.Value
}
