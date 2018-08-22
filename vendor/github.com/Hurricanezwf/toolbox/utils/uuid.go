package utils

import (
	"crypto/rand"
	"fmt"

	lib "github.com/satori/go.uuid"
)

func UUID() string {
	plain1, err := lib.NewV4()
	if err == nil {
		return plain1.String()
	}

	plan2 := make([]byte, 16)
	n, err := rand.Read(plan2)
	if err != nil || n != len(plan2) {
		return ""
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", plan2[0:4], plan2[4:6], plan2[6:8], plan2[8:10], plan2[10:])
}
