package types

func NewString(v string) *PB_String {
	return &PB_String{
		Value: v,
	}
}
