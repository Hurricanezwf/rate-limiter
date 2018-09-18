package ratelimiter

import "time"

// TransferInterface 传输层负责与集群建立连接并收发消息, 传输层自身维护与集群所有结点的TCP连接，并keepalive。
//
// 高可用要求:
// * 当与Leader出现网络隔离后，要求能通过Follower结点转发至Leader
// * 自动断线重连
type TransferInterface interface {
	// Open 打开传输层
	Open(conf *TransferConfig) error

	// Close 关闭传输层
	Close() error

	// PushBack 投递会话
	PushBack(sessionId uint32, requestBody []byte, timeout time.Duration) error

	// SetCallback 设置会话回调
	SetCallback(chan<- *TransferResponse)
}

type TransferResponse struct {
	SessionID uint32
	Error     error
	Response  []byte
}

type TransferConfig struct {
	// 服务器集群地址
	Cluster []string

	// KeepaliveInterval 心跳间隔
	KeepaliveInterval time.Duration
	// KeepaliveTimeout 心跳超时
	KeepaliveTimeout time.Duration

	// MaxRequestBacklog 最大请求积压
	MaxRequestBacklog int
}

func DefaultTransferConfig(cluster []string) *TransferConfig {
	return &TransferConfig{
		KeepaliveInterval: time.Second,
		KeepaliveTimeout:  5 * time.Second,
		MaxRequestBacklog: 64,
	}
}
