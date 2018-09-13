package tools

import (
	"testing"

	"github.com/Hurricanezwf/rate-limiter/proto"
)

func TestEncodeMessage(t *testing.T) {
	b, err := EncodeMessage(0x01, &proto.APIReturnReq{
		RCID:     "aaa",
		ClientID: []byte("zwf"),
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	t.Logf("%#v\n", b)
}
