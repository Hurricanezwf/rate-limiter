package encoding

const (
	VTypeBytes        byte = 0x01
	VTypeString       byte = 0x02
	VTypeInt8         byte = 0x03
	VTypeInt16        byte = 0x04
	VTypeInt32        byte = 0x05
	VTypeInt64        byte = 0x06
	VTypeUint8        byte = 0x07
	VTypeUint16       byte = 0x08
	VTypeUint32       byte = 0x09
	VTypeUint64       byte = 0x0a
	VTypeMap          byte = 0x0b
	VTypeQueue        byte = 0x0c
	VTypeBorrowRecord byte = 0x30
	VTypeLimiterMeta  byte = 0x31
)
