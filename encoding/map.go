package encoding

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Map struct {
	v map[string]Serializer
}

func NewMap() *Map {
	return &Map{
		v: make(map[string]Serializer),
	}
}

// Encode encode map to format which used to binary persist
// Encoded Format:
// > [1 byte]  data type
// > [4 bytes] pair's count of the map
// > --------------------------------------------
// > [1 byte]  type of the key
// > [4 bytes] byte count of key's encoded result
// > [N bytes] content of encoded key
// > [1 byte]  type of value
// > [4 bytes] byte count of value's encoded result
// > [N bytes] content of encoded value
// > --------------------------------------------
// > ...
// > ...
// > ...
func (m *Map) Encode() ([]byte, error) {
	header := make([]byte, 5)

	// 数据类型
	header[0] = VTypeMap

	// Map中元素个数
	binary.BigEndian.PutUint32(header[1:5], uint32(len(m.v)))

	// 循环写入map数据
	buf := bytes.NewBuffer(header)
	buf.Grow(10240)

	for k, v := range m.v {
		// encode key (string)
		kStr := NewString(k)
		kBytes, err := kStr.Encode()
		if err != nil {
			return nil, fmt.Errorf("Encode map's key(%s) failed, %v", k, err)
		}
		buf.Write(kBytes)

		// encode value
		vBytes, err := v.Encode()
		if err != nil {
			return nil, fmt.Errorf("Encode map's value failed, %v.  key=%s", err, k)
		}
		buf.Write(vBytes)
	}
	return buf.Bytes(), nil
}

func (m *Map) Decode(b []byte) ([]byte, error) {
	if len(b) < 5 {
		return nil, errors.New("Bad encoded format for map, too short")
	}

	// 数据类型
	if vType := b[0]; vType != VTypeMap {
		return nil, fmt.Errorf("Bad encoded format for map, VType(%#x) don't match %#x", vType, VTypeMap)
	}

	// map中元素个数
	mLen := binary.BigEndian.Uint32(b[1:5])

	// 解析map数据
	m.v = make(map[string]Serializer)
	b = b[5:]
	for i := uint32(0); i < mLen; i++ {
		// 判定长度是否足够记录KV的类型
		if len(b) < 2 {
			return nil, errors.New("Bad encoded format for map' KV, too short")
		}

		// 解析Key
		kType := b[0]
		key, err := SerializerFactory(kType)
		if err != nil {
			return nil, fmt.Errorf("Get serializer for map key failed, %v", err)
		}
		b, err = key.Decode(b)
		if err != nil {
			return nil, fmt.Errorf("Decode map key failed, %v", err)
		}
		keyStr := key.(*String).Value()

		// 解析Value
		//fmt.Printf("(%d) Before: %#v\n", i, b)
		vType := b[0]
		v, err := SerializerFactory(vType)
		if err != nil {
			return nil, fmt.Errorf("Get serializer for map value failed, %v", err)
		}
		b, err = v.Decode(b)
		if err != nil {
			return nil, fmt.Errorf("Decode map value failed, %v", err)
		}
		//fmt.Printf("(%d) After: %#v\n", i, b)

		// 添加进map
		m.Set(keyStr, v)
	}
	return b, nil
}

func (m *Map) Get(k string) (Serializer, bool) {
	vv, ok := m.v[k]
	return vv, ok
}

func (m *Map) Set(k string, v Serializer) {
	m.v[k] = v
}

// Range 遍历map，上层禁止关闭*KVPair channel
func (m *Map) Range(quit <-chan struct{}) <-chan *KVPair {
	if quit == nil {
		panic("quit is nil")
	}

	ch := make(chan *KVPair)

	go func() {
		defer close(ch)

		for kk, vv := range m.v {
			select {
			case ch <- &KVPair{K: kk, V: vv}:
				continue
			case <-quit:
				return
			}
		}
	}()

	return ch
}

func (m *Map) Delete(k string) {
	delete(m.v, k)
}

type KVPair struct {
	K string
	V Serializer
}
