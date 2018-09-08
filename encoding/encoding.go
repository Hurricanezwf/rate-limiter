package encoding

import "encoding/base64"

type Serializer interface {
	// Encode encode object to []byte
	Encode() ([]byte, error)

	// Decode decode from []byte to object
	// If decode success, byte sequeue will be removed from input []byte and then return
	Decode([]byte) ([]byte, error)
}

func BytesToString(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

func StringToBytes(v string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(v)
}
