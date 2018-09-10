package types

func NewInt64(v int64) *PB_Int64 {
	return &PB_Int64{
		Value: v,
	}
}
