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

func TestInt64(t *testing.T) {
	i := NewInt64(int64(89))
	i.Decode(nil)
	i.Incr(1)
	t.Logf("%v\n", i.Value())
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

func TestMap(t *testing.T) {
	m := NewMap()
	m.Set("name", NewString("lkx"))
	m.Set("name2", NewString("bug"))
	m.Set("name3r", NewString("lkx"))
	m.Decode(nil)

	v, ok := m.Get("name")
	if !ok {
		t.Fatal("Not found")
	} else {
		t.Logf("name:%v\n\n----------------------------------\n", v.(*String).Value())
	}

	ch := make(chan *KVPair)
	go m.Range(ch)

	for {
		pair, ok := <-ch
		if !ok {
			break
		}
		t.Logf("name:%v\n", pair.V.(*String).Value())
	}
}