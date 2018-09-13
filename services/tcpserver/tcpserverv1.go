package tcpserver

import (
	"net"

	"github.com/Hurricanezwf/rate-limiter/limiter"
)

type tcpserverv1 struct {
	// 服务器配置
	conf *Config

	// TCP监听器
	listener *net.TCPListener

	// 连接管理器
	mgr map[net.Conn]struct{}

	// Limiter实例
	limiter limiter.Interface
}

func newTCPServerV1() Interface {
	return &tcpserverv1{
		mgr: make(map[net.Conn]struct{}),
	}
}

// Open 打开TCP服务器
// 注意：limiter必须是处于打开状态
func (s *tcpserverv1) Open(conf *Config, l limiter.Interface) error {
	// TODO:
	return nil
}

func (s *tcpserverv1) Close() error {
	// TODO:
	return nil
}

func (s *tcpserverv1) ValidateConfig(conf *Config) error {
	// TODO
	return nil
}
