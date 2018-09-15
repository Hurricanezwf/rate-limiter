package limiter

import (
	"fmt"
	"testing"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/proto"
)

var (
	rcType        = []byte("zwf")
	clientID      = []byte("unitest")
	quota         = uint32(100000)
	resetInterval = int64(60)
	expire        = int64(100)
)

func BenchmarkBorrow(b *testing.B) {
	var err error

	g.ConfPath = "/home/zwf/zhouwenfeng/GoProj/src/github.com/Hurricanezwf/rate-limiter/conf/config.yaml"
	if err = g.Init(); err != nil {
		b.Fatal(err.Error())
	}

	var l = Default()
	if err = l.Open(); err != nil {
		b.Fatal(err.Error())
	}

	registRP := l.RegistQuota(&proto.APIRegistQuotaReq{
		RCType:        rcType,
		Quota:         quota,
		ResetInterval: resetInterval,
	})
	if registRP.Code != 0 {
		b.Fatal(registRP.Msg)
	}

	for i := 0; i < b.N; i++ {
		borrowRP := l.Borrow(&proto.APIBorrowReq{
			RCType:   rcType,
			ClientID: clientID,
			Expire:   expire,
		})
		if borrowRP.Code != 0 {
			fmt.Println(borrowRP.Msg)
			continue
		}

		returnRP := l.Return(&proto.APIReturnReq{
			RCID:     borrowRP.RCID,
			ClientID: clientID,
		})
		if returnRP.Code != 0 {
			fmt.Println(returnRP.Msg)
			continue
		}
	}
}
