package encoding

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type Uint32 struct {
	v uint32
}

func NewUint32(v uint32) *Uint32 {
	return &Uint32{v: v}
}

// Encode encode uint32 to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [4 bytes] data content
func (u *Uint32) Encode() ([]byte, error) {
	encoded := make([]byte, 5)
	encoded[0] = VTypeUint32
	binary.BigEndian.PutUint32(encoded[1:5], u.v)
	return encoded, nil
}

func (u *Uint32) Decode(b []byte) ([]byte, error) {
	// 校验头部信息
	if len(b) < 5 {
		return nil, errors.New("Bad encoded format for uint32")
	}

	// 校验数据类型
	if vType := b[0]; vType != VTypeUint32 {
		return nil, fmt.Errorf("Bad encoded format for uint32, VType(%x) don't match %x", vType, VTypeUint32)
	}

	// 解析数据
	u.v = binary.BigEndian.Uint32(b[1:5])

	return b[5:], nil
}

func (u *Uint32) Value() uint32 {
	return u.v
}

func (u *Uint32) Incr(delta uint32) {
	u.v += delta
}

func (u *Uint32) Decr(delta uint32) {
	u.v -= delta
}
