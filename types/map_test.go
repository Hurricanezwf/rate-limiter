package types

import (
	"testing"
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
