package ratelimiter

import (
	"net"
	"time"
)

// TransferInterface 传输层负责与集群建立连接并收发消息, 传输层自身维护与集群所有结点的TCP连接，并keepalive。
//
// 高可用要求:
// * 当与Leader出现网络隔离后，要求能通过Follower结点转发至Leader
// * 自动断线重连
type TransferInterface interface {
	// Open 打开传输层, 每次使用前必须先打开
	Open(conf *TransferConfig) error

	// Close 关闭传输层
	Close() error

	// PushBack 投递会话
	// 异步传输会话，如果再超时时间内未成功投递进传输层队列，则返回超时
	PushBack(sessionId uint32, requestBody []byte, timeout time.Duration) error

	// SetReadCallback 设置会话回调
	// 当传输层从socket中读取到一次会话的内容后，将通过这个通道传递给SessionMgr
	SetReadCallback(chan<- *TransferRead)
}

type TransferRead struct {
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

type transferV1 struct {
	// conf transfer的配置
	conf *TransferConfig

	// connMgr 连接管理器
	connMgr map[string]*net.TCPConn

	// 会话回调通道
}

func newTransferV1() *TransferInterface {
	return &transferV1{}
}

// Open 启动传输层
//
// 过程：
// * 分别与集群结点建立TCP连接
// * 每个连接分配两个协程分别用于读写
// * 启动keepalive
func (t *transferV1) Open(conf *TransferConfig) error {
	// TODO
	return nil
}

// Close 关闭传输层
func (t *trasferV1) Close() error {
	// TODO
	return nil
}

// PushBack 投递会话到传输层队列
func (t *transferV1) PushBack(sessionId uint32, requestBody []byte, timeout time.Duration) error {
	// TODO
	return nil
}

func (t *transferV1) SetReadCallback(chan *TransferRead) {
	// TODO
}
