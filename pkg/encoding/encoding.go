package encoding

import (
	"bufio"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/Hurricanezwf/rate-limiter/pkg/buffers"
	"github.com/golang/protobuf/proto"
)

func BytesToString(b []byte) string {
	return base64.RawStdEncoding.EncodeToString(b)
}

func StringToBytes(v string) ([]byte, error) {
	return base64.RawStdEncoding.DecodeString(v)
}

// 通信协议魔数
var msgMagicNumber byte = 0x01

// EncodeMsg 自定义协议格式
// [1 Byte ] Magic Number
// [1 Byte ] Result Code
// [1 Byte ] Action Type
// [4 Bytes] Sequence Number
// [4 Bytes] Data Len
// [N Bytes] Data Content
func EncodeMsg(action, resultCode byte, seq uint32, message proto.Message) ([]byte, error) {
	var err error
	var buf = buffers.GetProtoBuffer()
	var msg = buffers.GetBytesBuffer()

	// Header部分，共11字节
	if err = msg.WriteByte(msgMagicNumber); err != nil {
		goto FINISH
	}
	if err = msg.WriteByte(resultCode); err != nil {
		goto FINISH
	}
	if err = msg.WriteByte(action); err != nil {
		goto FINISH
	}
	if err = binary.Write(msg, binary.BigEndian, seq); err != nil {
		goto FINISH
	}
	if err = binary.Write(msg, binary.BigEndian, uint32(proto.Size(message))); err != nil {
		goto FINISH
	}
	// Payload部分
	if message != nil {
		if err = buf.Marshal(message); err != nil {
			goto FINISH
		}
		if _, err = msg.Write(buf.Bytes()); err != nil {
			goto FINISH
		}
	}

FINISH:
	b := msg.Bytes()
	buffers.PutProtoBuffer(buf)
	buffers.PutBytesBuffer(msg)
	return b, err
}

// DecodeMsg 解码消息
func DecodeMsg(reader *bufio.Reader) (action, resultCode byte, seq uint32, msgBody []byte, err error) {
	var (
		magicNumber byte
		dataLen     uint32
	)

	// 魔数
	magicNumber, err = reader.ReadByte()
	if err != nil {
		return
	}
	if magicNumber != msgMagicNumber {
		err = fmt.Errorf("MagicNumer %#v not match", magicNumber)
		return
	}

	// Result Code
	resultCode, err = reader.ReadByte()
	if err != nil {
		err = fmt.Errorf("Read result code failed, %v", err)
		return
	}

	// Action
	action, err = reader.ReadByte()
	if err != nil {
		err = fmt.Errorf("Read action failed, %v", err)
		return
	}

	// 序列号
	err = binary.Read(reader, binary.BigEndian, &seq)
	if err != nil {
		err = fmt.Errorf("Read sequence number failed, %v", err)
		return
	}

	// 数据长度
	err = binary.Read(reader, binary.BigEndian, &dataLen)
	if err != nil {
		err = fmt.Errorf("Read msg length failed, %v", err)
		return
	}

	// 数据内容, 这里msgBytes务必重新分配空间，因为它会被作为参数传递进Event
	if dataLen > 0 {
		msgBody = make([]byte, dataLen)
		_, err = reader.Read(msgBody)
		if err != nil && err != io.EOF {
			err = fmt.Errorf("Read msg failed, %v", err)
			return
		}
	}

	return action, resultCode, seq, msgBody, err
}
