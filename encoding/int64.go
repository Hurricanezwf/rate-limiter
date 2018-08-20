package encoding

type Int64 struct {
	v int64
}

func NewInt64(v int64) *Int64 {
	return &Int64{v: v}
}

func (i *Int64) Encode() ([]byte, error) {
	// TODO
	return nil, nil
}

func (i *Int64) Decode(b []byte) ([]byte, error) {
	// TODO
	i.v = int64(12345)
	return nil, nil
}

func (i *Int64) Value() int64 {
	return i.v
}

func (i *Int64) Incr(delta int64) {
	i.v += delta
}
