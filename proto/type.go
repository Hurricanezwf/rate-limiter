package proto

import "encoding/hex"

// Action类型
type ActionType byte

// 资源类型ID
type ResourceTypeID []byte

func (id ResourceTypeID) String() string {
	return hex.EncodeToString(id)
}

func (id ResourceTypeID) Len() int {
	return len(id)
}

//func (id ResourceTypeID) Encode() ([]byte, error) {
//	if len(id) > math.MaxUint8 {
//		return nil, errors.New("ResourceTypeID overflow")
//	}
//	encoded := make([]byte, len(id)+2)
//	encoded[0] = ValueTypeResourceTypeID
//	encoded[1] = uint8(len(id))
//	if len(id) > 0 {
//		copy(encoded[2:], id[:])
//	}
//	return encoded, nil
//}
//
//func (id ResourceTypeID) Decode(b []byte) ([]byte, error) {
//	if len(b) < 2 {
//		return b, errors.New("Bad ResourceTypeID byte sequence")
//	}
//	if vType := b[0]; vType != ValueTypeResourceTypeID {
//		return b, errors.New("Value type not match")
//	}
//
//	rcTypeIdLen := int(b[1])
//	if len(b) > 2 {
//		id = ResourceTypeID(b[2 : 2+rcTypeIdLen])
//		b = b[2+rcTypeIdLen:]
//	} else {
//		b = nil
//	}
//	return b, nil
//}

// 资源ID
type ResourceID string

func (id ResourceID) Len() int {
	return len(id)
}

//func (id ResourceID) Encode() ([]byte, error) {
//	if len(id) > math.MaxUint8 {
//		return nil, errors.New("ResourceID overflow")
//	}
//	encoded := make([]byte, len(id)+2)
//	encoded[0] = ValueTypeResourceID
//	encoded[1] = byte(len(id))
//	if len(id) > 0 {
//		copy(encoded[2:], []byte(id)[:])
//	}
//	return encoded, nil
//}
//
//func (id ResourceID) Decode(b []byte) ([]byte, error) {
//	if len(b) < 2 {
//		return b, errors.New("Bad ResouceID byte sequence")
//	}
//	if vType := b[0]; vType != ValueTypeResourceID {
//		return b, errors.New("Valut type not match")
//	}
//
//	rcIdLen := int(b[1])
//	if len(b) > 2 {
//		id = ResourceID(b[2 : 2+rcIdLen])
//		b = b[2+rcIdLen:]
//	} else {
//		b = nil
//	}
//	return b, nil
//}

// 客户端ID
type ClientID []byte

func (id ClientID) String() string {
	return hex.EncodeToString(id)
}

func (id ClientID) Len() int {
	return len(id)
}

//func (id ClientID) Encode() ([]byte, error) {
//	if len(id) > math.MaxUint8 {
//		return nil, errors.New("ClientID overflow")
//	}
//	encoded := make([]byte, len(id)+2)
//	encoded[0] = ValueTypeClientID
//	encoded[1] = byte(len(id))
//	if len(id) > 0 {
//		copy(encoded[2:], id[:])
//	}
//	return encoded, nil
//}
//
//func (id ClientID) Decode(b []byte) ([]byte, error) {
//	if len(b) < 2 {
//		return b, errors.New("Bad ClientID byte sequence")
//	}
//	if vType := b[0]; vType != ValueTypeClientID {
//		return b, errors.New("Value type not match")
//	}
//
//	clientIdLen := int(b[1])
//	if len(b) > 2 {
//		id = ClientID(b[2 : 2+clientIdLen])
//		b = b[2+clientIdLen:]
//	} else {
//		b = nil
//	}
//	return b, nil
//}
