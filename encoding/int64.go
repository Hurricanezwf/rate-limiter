package encoding

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Int64 struct {
	v int64
}

func NewInt64(v int64) *Int64 {
	return &Int64{v: v}
}

// Encode encode int64 to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [8 bytes] data content
func (i *Int64) Encode() ([]byte, error) {
	encoded := make([]byte, 9)
	encoded[0] = VTypeInt64
	binary.BigEndian.PutUint64(encoded[1:9], uint64(i.v))
	return encoded, nil
}

func (i *Int64) Decode(b []byte) ([]byte, error) {
	// 校验头部信息
	if len(b) < 9 {
		return nil, errors.New("Bad encoded format for int64")
	}

	// 校验数据类型
	if vType := b[0]; vType != VTypeInt64 {
		return nil, fmt.Errorf("Bad encoded format for int64, VType(%x) don't match %x", vType, VTypeInt64)
	}

	// 解析数据
	i.v = int64(binary.BigEndian.Uint64(b[1:9]))

	return b[9:], nil
}

func (i *Int64) Value() int64 {
	return i.v
}

func (i *Int64) Incr(delta int64) {
	i.v += delta
}
