package types

func NewFloat64(v float64) *PB_Float64 {
	return &PB_Float64{
		Value: v,
	}
}
