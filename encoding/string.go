package encoding

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type String struct {
	v string
}

func NewString(v string) *String {
	return &String{v: v}
}

// Encode encode string to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [4 bytes] data len
// > [N bytes] data content
func (s *String) Encode() ([]byte, error) {
	encoded := make([]byte, 5+len(s.v))
	encoded[0] = VTypeString
	binary.BigEndian.PutUint32(encoded[1:5], uint32(len(s.v)))
	copy(encoded[5:], []byte(s.v)[:])
	return encoded, nil
}

func (s *String) Decode(b []byte) ([]byte, error) {
	// 校验头部信息
	if len(b) < 5 {
		return nil, errors.New("Bad encoded format for string, too short")
	}

	// 校验数据类型
	if vType := b[0]; vType != VTypeString {
		return nil, fmt.Errorf("Bad encoded format for string, VType(%x) don't match %x", vType, VTypeString)
	}

	// 校验数据长度
	vLen := binary.BigEndian.Uint32(b[1:5])
	if uint32(len(b)) < 5+vLen {
		return nil, errors.New("Bad encoded format for string, too short")
	}

	// 解析数据
	s.v = string(b[5 : 5+vLen])

	return b[5+vLen:], nil

}

func (s *String) Value() string {
	return s.v
}
