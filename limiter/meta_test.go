package limiter

import (
	"crypto/md5"
	"testing"

	. "github.com/Hurricanezwf/rate-limiter/proto"
)

var (
	tId   ResourceTypeID
	cId   ClientID
	quota = 4
)

func TestLimiterMeta(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = ResourceTypeID(tIdMd5[:])
	cId = ClientID(cIdMd5[:])

	m := NewLimiterMeta(tId, quota)

	for i := 0; i < 5; i++ {
		rcId, err := m.Borrow(tId, cId, 10)
		if err != nil {
			t.Fatal(err.Error())
		} else {
			t.Logf("Get RC %s", rcId)
		}
	}
}
