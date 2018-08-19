package limiter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	. "github.com/Hurricanezwf/rate-limiter/proto"
)

func ValueTypeMatch(vType byte, v interface{}) (match bool) {
	switch vType {
	case ValueTypeString:
		_, match = v.(string)
	case ValueTypeUint8:
		_, match = v.(uint8)
	case ValueTypeUint32:
		_, match = v.(uint32)
	case ValueTypeUint64:
		_, match = v.(uint64)
	case ValueTypeResourceID:
		_, match = v.(ResourceID)
	case ValueTypeResourceTypeID:
		_, match = v.(ResourceTypeID)
	case ValueTypeClientID:
		_, match = v.(ClientID)
	case ValueTypeBorrowRecord:
		_, match = v.(borrowRecord)
	default:
		panic(fmt.Sprintf("Unknown valut type %v", vType))
	}
	return match
}

func MakeResourceID(tId ResourceTypeID, idx uint32) ResourceID {
	return ResourceID(fmt.Sprintf("%s_rc#%d", tId.String(), idx))
}

func ResolveResourceID(rcId ResourceID) (ResourceTypeID, error) {
	arr := strings.Split(string(rcId), "_")
	if len(arr) != 2 {
		return nil, errors.New("Bad ResourceID")
	}

	tId, err := hex.DecodeString(arr[0])
	if err != nil {
		return nil, err
	}
	return ResourceTypeID(tId), nil
}
