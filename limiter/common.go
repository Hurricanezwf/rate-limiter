package limiter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

//func ValueTypeMatch(vType byte, v interface{}) (match bool) {
//	switch vType {
//	case ValueTypeString:
//		_, match = v.(string)
//	case ValueTypeUint8:
//		_, match = v.(uint8)
//	case ValueTypeUint32:
//		_, match = v.(uint32)
//	case ValueTypeUint64:
//		_, match = v.(uint64)
//	case ValueTypeResourceID:
//		_, match = v.(ResourceID)
//	case ValueTypeResourceTypeID:
//		_, match = v.(ResourceTypeID)
//	case ValueTypeClientID:
//		_, match = v.(ClientID)
//	case ValueTypeBorrowRecord:
//		_, match = v.(borrowRecord)
//	default:
//		panic(fmt.Sprintf("Unknown valut type %v", vType))
//	}
//	return match
//}

func MakeResourceID(rcTypeId []byte, idx uint32) string {
	return fmt.Sprintf("%x_rc#%d", rcTypeId, idx)
}

func ResolveResourceID(rcId string) (rcTypeId []byte, err error) {
	arr := strings.Split(rcId, "_")
	if len(arr) != 2 {
		return nil, errors.New("Bad ResourceID")
	}

	return hex.DecodeString(arr[0])
}
