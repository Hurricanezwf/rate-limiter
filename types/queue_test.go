package types

import (
	"fmt"
	"testing"
)

func TestQueue(t *testing.T) {
	q1 := NewQueue()
	t.Logf("Queue1 Len:%d\n", q1.Len())

	q1.PushBack(NewString("hello"))
	t.Logf("Queue1 Len:%d\n", q1.Len())

	q1.PushBack(NewString("world"))
	t.Logf("Queue1 Len:%d\n", q1.Len())

	q2 := NewQueue()
	q2.PushBackList(q1)
	t.Logf("Queue2 Len:%d\n", q2.Len())

	e := q1.PopFront()
	str, err := AnyToString(e.Value)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%s poped from q1, len:%d\n", str.Value, q1.Len())

	front := q1.Front()
	str, err = AnyToString(front.Value)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("Front is %s\n", str.Value)

	back := q1.Back()
	str, err = AnyToString(back.Value)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("Back is %s\n", str.Value)

	q1.Remove(back)
	t.Logf("Queue1 Len:%d\n", q1.Len())
}
func TestQueuePushBack(t *testing.T) {
	q := NewQueue()

	for i := 0; i < 10; i++ {
		str := NewString(fmt.Sprintf("ID#%d", i))
		q.PushBack(str)
	}

	for itr := q.Front(); itr != nil; itr = itr.Next {
		str := NewString("")
		UnmarshalAny(itr.Value, str)
		t.Logf("%s\n", str.Value)
	}
}
