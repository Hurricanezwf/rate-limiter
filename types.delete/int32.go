package types

func NewInt32(v int32) *PB_Int32 {
	return &PB_Int32{
		Value: v,
	}
}
