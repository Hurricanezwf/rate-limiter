package ratelimiter

import (
	"fmt"
	"net"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// RateLimiter客户端代理库，走TCP协议
type TCPClient struct {
	mutex *sync.RWMutex

	// 配置
	config *ClientConfig

	// 客户端ID，每个实例不重复
	clientId []byte

	// 使用过的资源类型, 这里使用base64存储
	usedRCType map[string]struct{}

	// TCP Client
	tcpClient *net.TCPConn

	// 控制客户端关闭
	stopC chan struct{}
}

func NewTCPClient(config *ClientConfig) (Interface, error) {
	// 校验配置
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	// 生成随机ClientID
	clientId, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("Generate uuid for client failed, %v", err)
	}

	c := TCPClient{
		mutex:      &sync.RWMutex{},
		clientId:   clientId.Bytes(),
		config:     config,
		usedRCType: make(map[string]struct{}),
		tcpClient:  *net.TCPAddr,
	}
}
