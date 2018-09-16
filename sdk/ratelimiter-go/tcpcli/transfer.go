package ratelimiter

import "time"

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
