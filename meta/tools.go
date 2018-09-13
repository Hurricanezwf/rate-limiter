package meta

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Hurricanezwf/rate-limiter/tools"
)

func MakeResourceID(rcTypeId []byte, idx uint32) string {
	return fmt.Sprintf("%s_rc#%d", tools.BytesToString(rcTypeId), idx)
}

func ResolveResourceID(rcId string) (rcTypeId []byte, err error) {
	arr := strings.Split(rcId, "_")
	if len(arr) != 2 {
		return nil, errors.New("Bad ResourceID")
	}

	return tools.StringToBytes(arr[0])
}
