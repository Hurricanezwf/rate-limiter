package tools

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"sync"

	"github.com/gogo/protobuf/proto"
)

// protobuffer对象池
var protoBufferPool = sync.Pool{
	New: func() interface{} {
		return proto.NewBuffer(make([]byte, 256))
	},
}

// bytes buffer对象池
var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 256))
	},
}

// bufio.Reader对象池
var bufioReaderPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReader(nil)
	},
}

func BytesToString(b []byte) string {
	return base64.RawStdEncoding.EncodeToString(b)
}

func StringToBytes(v string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(v)
}

func EncodeMsg(magicNumber, action byte, message proto.Message) ([]byte, error) {
	var err error
	var buf = GetProtoBuffer()
	var msg = GetBytesBuffer()

	if err = buf.Marshal(message); err != nil {
		goto FINISH
	}
	if err = msg.WriteByte(magicNumber); err != nil {
		goto FINISH
	}
	if err = msg.WriteByte(action); err != nil {
		goto FINISH
	}
	if err = binary.Write(msg, binary.BigEndian, uint32(proto.Size(message))); err != nil {
		goto FINISH
	}
	if _, err = msg.Write(buf.Bytes()); err != nil {
		goto FINISH
	}

FINISH:
	b := msg.Bytes()
	PutProtoBuffer(buf)
	PutBytesBuffer(msg)
	return b, err
}

func GetProtoBuffer() *proto.Buffer {
	return protoBufferPool.Get().(*proto.Buffer)
}

func PutProtoBuffer(buf *proto.Buffer) {
	buf.Reset()
	protoBufferPool.Put(buf)
}

func GetBytesBuffer() *bytes.Buffer {
	return bytesBufferPool.Get().(*bytes.Buffer)
}

func PutBytesBuffer(buf *bytes.Buffer) {
	buf.Reset()
	bytesBufferPool.Put(buf)
}

func GetBufioReader() *bufio.Reader {
	return bufioReaderPool.Get().(*bufio.Reader)
}
