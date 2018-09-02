package types

type Interface interface {
	// Encode 将类型序列化
	Encode() ([]byte, error)

	// Decode 将类型反序列化
}
