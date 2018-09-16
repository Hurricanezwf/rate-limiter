package ratelimiter

import (
	"time"

	"github.com/gogo/protobuf/proto"
)

type SessionInterface interface {
	// ID 获取session id
	ID() uint32

	// Do 投递请求进入发送队列并监听响应
	// 此函数是同步返回的
	Do(action byte, msg proto.Message, timeout time.Duration) (rsp []byte, err error)

	// Done 由mgr调用告知session处理请求结果
	Done(rsp *TransferResponse)

	// Reset 重置会话
	Reset()
}

func DefaultSession() SessionInterface {
	return newSession()
}

type Session struct {
	// sessionID 会话ID, 用于识别上下文
	sessionID uint32

	// readCh 接收传输层响应结果
	readCh chan *TransferResponse
}

func newSession() SessionInterface {
	return &Session{
		sessionID: 0,
		readCh:    make(chan *TransferResponse, 1),
	}
}
