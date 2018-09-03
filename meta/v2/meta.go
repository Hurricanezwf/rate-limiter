package metav2

import (
	"github.com/Hurricanezwf/rate-limiter/meta"
	"github.com/Hurricanezwf/rate-limiter/types"
)

func init() {
	meta.RegistMetaBuilder("v2", newLimiterV2)
}

func newLimiterV2(rcTypeId []byte, quota uint32) meta.LimiterMeta {
	return &LimiterMetaV2{
		RcTypeId:  types.NewBytes(rcTypeId),
		Quota:     types.NewUint32(quota),
		CanBorrow: types.NewQueue(),
		Recycled:  types.NewQueue(),
		Used:      types.NewMap(),
	}
}

func (m *LimiterMetaV2) Borrow(clientId []byte, expire int64) (string, error) {

}

func (m *LimiterMetaV2) Return(clientId []byte, rcId string) error {

}

func (m *LimiterMetaV2) ReturnAll(clientId []byte) (int, error) {

}

func (m *LimiterMetaV2) Recycle() {

}
