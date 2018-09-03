package types

func NewUint64(v uint64) *PB_Uint64 {
	return &PB_Uint64{
		Value: v,
	}
}
