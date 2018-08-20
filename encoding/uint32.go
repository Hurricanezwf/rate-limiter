package encoding

type Uint32 struct {
	v uint32
}

func NewUint32(v uint32) *Uint32 {
	return &Uint32{v: v}
}

func (u *Uint32) Encode() ([]byte, error) {
	// TODO
	return nil, nil
}

func (u *Uint32) Decode(b []byte) ([]byte, error) {
	// TODO:
	u.v = uint32(100)
	return nil, nil
}

func (u *Uint32) Value() uint32 {
	return u.v
}
