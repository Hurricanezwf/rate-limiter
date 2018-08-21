package encoding

import "fmt"

var generators = make(map[byte]SerializerGenerator)

// Serializer生成器
type SerializerGenerator func() Serializer

// 注册Serializer生成器
func RegistSerializerGenerator(vType byte, f SerializerGenerator) error {
	if f == nil {
		return fmt.Errorf("SerializerGenerator for %#x is nil", vType)
	}
	if _, existed := generators[vType]; existed {
		return fmt.Errorf("SerializerGenerator for %#x had been existed", vType)
	}
	generators[vType] = f
	return nil
}

// Serializer工厂函数，生成Serializer实例
func SerializerFactory(vType byte) (Serializer, error) {
	f, existed := generators[vType]
	if !existed {
		return nil, fmt.Errorf("Not found for %#x", vType)
	}
	return f(), nil
}
