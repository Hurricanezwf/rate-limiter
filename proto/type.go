package proto

import (
	"encoding/hex"
	"errors"
	"math"
)

// Action类型
type ActionType uint8

// 资源类型ID
type ResourceTypeID []byte

func (id ResourceTypeID) String() string {
	return hex.EncodeToString(id)
}

func (id ResourceTypeID) Len() int {
	return len(id)
}

func (id ResourceTypeID) Encode() ([]byte, error) {
	if len(id) > math.MaxUint8 {
		return nil, errors.New("ResourceTypeID overflow")
	}
	encoded := make([]byte, len(id)+1)
	encoded[0] = uint8(len(id))
	if len(id) > 0 {
		copy(encoded[1:], id[:])
	}
	return encoded, nil
}

func (id ResourceTypeID) Decode(b []byte) ([]byte, error) {
	if len(b) < 1 {
		return b, errors.New("Bad ResourceTypeID byte sequence")
	}
	rcTypeIdLen := int(b[0])
	id = ResourceTypeID(b[1 : 1+rcTypeIdLen])
	return b[1+rcTypeIdLen:], nil
}

// 资源ID
type ResourceID string

func (id ResourceID) Len() int {
	return len(id)
}

func (id ResourceID) Encode() ([]byte, error) {
	if len(id) > math.MaxUint8 {
		return nil, errors.New("ResourceID overflow")
	}
	encoded := make([]byte, len(id)+1)
	encoded[0] = uint8(len(id))
	if len(id) > 0 {
		copy(encoded[1:], []byte(id)[:])
	}
	return encoded, nil
}

func (id ResourceID) Decode(b []byte) ([]byte, error) {
	if len(b) < 1 {
		return b, errors.New("Bad ResouceID byte sequence")
	}
	rcIdLen := int(b[0])
	id = ResourceID(b[1 : 1+rcIdLen])
	return b[1+rcIdLen:], nil
}

// 客户端ID
type ClientID []byte

func (id ClientID) String() string {
	return hex.EncodeToString(id)
}

func (id ClientID) Len() int {
	return len(id)
}

func (id ClientID) Encode() ([]byte, error) {
	if len(id) > math.MaxUint8 {
		return nil, errors.New("ClientID overflow")
	}
	encoded := make([]byte, len(id)+1)
	encoded[0] = uint8(len(id))
	if len(id) > 0 {
		copy(encoded[1:], id[:])
	}
	return encoded, nil
}

func (id ClientID) Decode(b []byte) ([]byte, error) {
	if len(b) < 1 {
		return b, errors.New("Bad ClientID byte sequence")
	}
	clientIdLen := int(b[0])
	id = ClientID(b[1 : 1+clientIdLen])
	return b[1+clientIdLen:], nil
}
