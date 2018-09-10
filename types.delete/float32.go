package types

func NewFloat32(v float32) *PB_Float32 {
	return &PB_Float32{
		Value: v,
	}
}
