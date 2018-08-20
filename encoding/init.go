package encoding

func init() {
	// 注册serializer生成器
	RegistSerializerGenerator(VTypeBytes, bytesSerializerGenerator)
	RegistSerializerGenerator(VTypeString, stringSerializerGenerator)
	RegistSerializerGenerator(VTypeInt64, int64SerializerGenerator)
	RegistSerializerGenerator(VTypeUint32, uint32SerializerGenerator)
	RegistSerializerGenerator(VTypeMap, mapSerializerGenerator)
	RegistSerializerGenerator(VTypeQueue, queueSerializerGenerator)
}

func bytesSerializerGenerator() Serializer {
	return NewBytes(nil)
}

func stringSerializerGenerator() Serializer {
	return NewString("")
}

func int64SerializerGenerator() Serializer {
	return NewInt64(0)
}

func uint32SerializerGenerator() Serializer {
	return NewUint32(0)
}

func mapSerializerGenerator() Serializer {
	return NewMap()
}

func queueSerializerGenerator() Serializer {
	return NewQueue()
}
