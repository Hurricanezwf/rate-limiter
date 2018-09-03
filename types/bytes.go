package types

func NewBytes(v []byte) *PB_Bytes {
	return &PB_Bytes{
		Value: v,
	}
}
