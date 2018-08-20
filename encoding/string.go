package encoding

type String struct {
	v string
}

func NewString(v string) *String {
	return &String{v: v}
}

func (s *String) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (s *String) Decode(b []byte) ([]byte, error) {
	// TODO
	s.v = "hello"
	return nil, nil
}

func (s *String) Value() string {
	return s.v
}
