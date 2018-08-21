package encoding

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
)

type Bytes struct {
	v []byte
}

func NewBytes(v []byte) *Bytes {
	return &Bytes{v: v}
}

// Encode encode byte slice to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [4 bytes] data len
// > [N bytes] data content
func (bt *Bytes) Encode() ([]byte, error) {
	encoded := make([]byte, 5+len(bt.v))
	encoded[0] = VTypeBytes
	binary.BigEndian.PutUint32(encoded[1:5], uint32(len(bt.v)))
	copy(encoded[5:], bt.v)
	return encoded, nil
}

func (bt *Bytes) Decode(b []byte) ([]byte, error) {
	// 校验头部信息
	if len(b) < 5 {
		return nil, errors.New("Bad encoded format for bytes, too short")
	}

	// 校验数据类型
	if vType := b[0]; vType != VTypeBytes {
		return nil, fmt.Errorf("Bad encoded format for bytes, VType(%#x) don't match %#x", vType, VTypeBytes)
	}

	// 校验数据长度
	vLen := binary.BigEndian.Uint32(b[1:5])
	if uint32(len(b)) < 5+vLen {
		return nil, errors.New("Bad encoded format for bytes, too short")
	}

	// 解析数据
	bt.v = b[5 : 5+vLen]

	return b[5+vLen:], nil
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
