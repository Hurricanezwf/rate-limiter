package types

import "testing"

func TestQueue(t *testing.T) {
	q := NewQueue()
	t.Logf("Queue Len:%d\n", q.Len())

	q.PushBack(PB_String())
}
