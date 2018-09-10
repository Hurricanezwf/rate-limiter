package types

func NewBool(v bool) *PB_Bool {
	return &PB_Bool{
		Value: v,
	}
}
