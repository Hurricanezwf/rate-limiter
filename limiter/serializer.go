package limiter

type Serializer interface {
	// Encode encode object to []byte
	Encode() ([]byte, error)

	// Decode decode from []byte to object
	Decode([]byte) error
}
