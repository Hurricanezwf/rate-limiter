package meta

import (
	"testing"

	"github.com/Hurricanezwf/rate-limiter/types"
	proto "github.com/golang/protobuf/proto"
)

func TestMetaMarshal(t *testing.T) {
	mgr := &PB_M{
		RcTypeId:  []byte("hello"),
		Quota:     100,
		CanBorrow: types.NewQueue(),
		Recycled:  types.NewQueue(),
		Used:      make(map[string]*types.PB_Queue),
	}

	mgr.CanBorrow.PushBack(types.NewString("rc#0"))
	//mgr.CanBorrow.PushBack(types.NewString("rc#1"))
	//mgr.Recycled.PushBack(types.NewString("rc#2"))
	//mgr.Used = map[string]*types.PB_Queue{
	//	"zwf": types.NewQueue(),
	//}
	//mgr.Used["zwf"].PushBack(types.NewString("rc#8"))

	m := &PB_Meta{
		Value: map[string]*PB_M{
			"hello": mgr,
		},
	}

	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err.Error())
	}
	_ = b
	//t.Logf("Len:%d, Bytes; %#v\n", len(b), b)
}
