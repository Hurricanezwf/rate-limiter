package types

import (
	"testing"

	proto "github.com/golang/protobuf/proto"
)

func TestMap(t *testing.T) {
	m := NewMap()
	m.Set("hello", NewString("World"))

	k := "hello"
	v := m.Get(k)
	if v == nil {
		t.Fatalf("No value found for %s\n", k)
	}

	str, err := AnyToString(v)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%s=%s, len:%d\n", k, str.Value, m.Len())

	m.Remove(k)
	t.Logf("len=%d\n", m.Len())
}

func TestMapMarshal(t *testing.T) {
	m := NewMap()
	m.Set("1", NewString("1"))

	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("Len:%d, Bytes: %#v\n", len(b), b)
}
