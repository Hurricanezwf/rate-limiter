package limiter

import "github.com/Hurricanezwf/rate-limiter/encoding"

func init() {
	encoding.RegistSerializerGenerator(encoding.VTypeBorrowRecord, borrowRecordSerializerGenerator)
}

func borrowRecordSerializerGenerator() encoding.Serializer {
	return &BorrowRecord{}
}
