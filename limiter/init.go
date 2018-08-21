package limiter

import "github.com/Hurricanezwf/rate-limiter/encoding"

func init() {
	encoding.RegistSerializerGenerator(encoding.VTypeBorrowRecord, borrowRecordSerializerGenerator)
	encoding.RegistSerializerGenerator(encoding.VTypeLimiterMeta, limiterMetaSerializerGenerator)
}

func borrowRecordSerializerGenerator() encoding.Serializer {
	return &BorrowRecord{}
}

func limiterMetaSerializerGenerator() encoding.Serializer {
	return &limiterMetaV1{}
}
