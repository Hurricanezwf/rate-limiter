package encoding

import "encoding/hex"

type Serializer interface {
	// Encode encode object to []byte
	Encode() ([]byte, error)

	// Decode decode from []byte to object
	// If decode success, byte sequeue will be removed from input []byte and then return
	Decode([]byte) ([]byte, error)
}

func BytesToStringHex(b []byte) string {
	return hex.EncodeToString(b)
}

func StringHexToBytes(v string) ([]byte, error) {
	return hex.DecodeString(v)
}
