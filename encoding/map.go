package encoding

type Map struct {
	v map[string]Serializer
}

func NewMap() *Map {
	return &Map{
		v: make(map[string]Serializer),
	}
}

func (m *Map) Encode() ([]byte, error) {
	// TODO:
	return nil, nil
}

func (m *Map) Decode(b []byte) ([]byte, error) {
	// TODO:
	m.v["name"] = NewString("zwf")
	return nil, nil
}

func (m *Map) Get(k string) (Serializer, bool) {
	vv, ok := m.v[k]
	return vv, ok
}

func (m *Map) Set(k string, v Serializer) {
	m.v[k] = v
}

func (m *Map) Range(ch chan<- *KVPair) {
	for kk, vv := range m.v {
		ch <- &KVPair{
			K: kk,
			V: vv,
		}
	}
	close(ch)
}

type KVPair struct {
	K string
	V Serializer
}
