package encoding

import "encoding/hex"

type Bytes struct {
	v []byte
}

func NewBytes(v []byte) *Bytes {
	return &Bytes{v: v}
}

func (bt *Bytes) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (bt *Bytes) Decode(b []byte) ([]byte, error) {
	// TODO
	bt.v = []byte("hello")
	return nil, nil
}

func (bt *Bytes) Set(idx int, bv byte) {
	bt.v[idx] = bv
}

func (bt *Bytes) Value() []byte {
	return bt.v
}

func (bt *Bytes) Hex() string {
	return hex.EncodeToString(bt.v)
}

func (bt *Bytes) Grow(n int) *Bytes {
	if cap(bt.v) >= n {
		return bt
	}

	v2 := make([]byte, 0, n)
	if len(bt.v) > 0 {
		v2 = append(v2, bt.v...)
	}
	bt.v = v2
	return bt
}

func (bt *Bytes) Len() int {
	return len(bt.v)
}

func (bt *Bytes) Cap() int {
	return cap(bt.v)
}
