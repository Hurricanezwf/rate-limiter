package encoding

import (
	"testing"
)

func TestString(t *testing.T) {
	s := NewString("world")
	s.Decode(nil)
	t.Logf("%s\n", s.Value())
}

func TestBytes(t *testing.T) {
	b := NewBytes([]byte("world"))
	b.Decode(nil)
	t.Logf("value:%v, len:%d, cap:%d\n", b.Value(), b.Len(), b.Cap())

	b.Grow(1024)
	t.Logf("value:%v, len:%d, cap:%d\n", b.Value(), b.Len(), b.Cap())
}

func TestUint32(t *testing.T) {
	b := NewUint32(uint32(2))
	b.Decode(nil)
	t.Logf("%v\n", b.Value())
}

func TestQueue(t *testing.T) {
	q := NewQueue()
	q.PushBack(NewString("zwf"))
	q.PushBack(NewString("lkx"))

	q.PopFront()

	q2 := NewQueue()
	q2.PushBack(NewString("bug"))
	q.PushBackQueue(q2)

	for e := q.Front(); e.IsNil() == false; e = e.Next() {
		v := e.Value().(*String)
		t.Logf("%s\n", v.Value())
	}
}
