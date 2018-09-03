package types

func NewUint32(v uint32) *PB_Uint32 {
	return &PB_Uint32{
		Value: v,
	}
}
