package encoding

import (
	"testing"
)

func TestString(t *testing.T) {
	s := NewString("world")

	// encode
	bt, err := s.Encode()
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Encoded: %#v\n", bt)
	}

	// decode
	if _, err = s.Decode(bt); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Decoded: %#v\n", s.Value())
	}
}

func TestBytes(t *testing.T) {
	b := NewBytes([]byte("world"))
	t.Logf("Origin: %#v, len:%d, cap:%d\n", b.Value(), b.Len(), b.Cap())

	// set
	b.Set(0, 0x00)
	t.Logf("After Set: %#v, len:%d, cap:%d\n", b.Value(), b.Len(), b.Cap())

	// grow
	b.Grow(1024)
	t.Logf("After Grow: %#v, len:%d, cap:%d\n", b.Value(), b.Len(), b.Cap())

	// encode
	bt, err := b.Encode()
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Encoded: %#v\n", bt)
	}

	// decode
	if _, err = b.Decode(bt); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Decoded: %#v\n", b.Value())
	}
}

func TestUint32(t *testing.T) {
	b := NewUint32(uint32(2))

	// incre
	b.Incr(uint32(1))
	t.Logf("After Decode: %v\n", b.Value())

	// encode
	bt, err := b.Encode()
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Encoded: %#v\n", bt)
	}

	// decode
	b.Incr(uint32(1))
	if _, err = b.Decode(bt); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Decoded: %v\n", b.Value())
	}
}

func TestInt64(t *testing.T) {
	i := NewInt64(int64(89))

	// incr
	i.Incr(1)
	t.Logf("After Incr: %v\n", i.Value())

	// encode
	b, err := i.Encode()
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Encoded: %#v\n", b)
	}

	// decode
	i.Incr(1)
	if _, err = i.Decode(b); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Decoded: %v\n", i.Value())
	}
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
