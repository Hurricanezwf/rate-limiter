package limiter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
)

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

func PathExists(p string) bool {
	if _, err := os.Lstat(p); err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
