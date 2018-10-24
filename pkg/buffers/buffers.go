package buffers

import (
	"bufio"
	"bytes"
	"sync"

	"github.com/golang/protobuf/proto"
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

func PutBufioReader(r *bufio.Reader) {
	bufioReaderPool.Put(r)
}
