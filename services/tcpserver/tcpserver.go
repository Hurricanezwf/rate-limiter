package tcpserver

import "github.com/Hurricanezwf/rate-limiter/limiter"

func init() {
	RegistBuilder("v1", newTCPServerV1)
}

type Config struct {
	// TCP服务器监听地址, 非0.0.0.0
	Listen string

	// 心跳间隔, 单位毫秒
	KeepAliveInterval int64

	// 多久收不到心跳则关闭连接，单位毫秒
	Deadline int64
}

type Interface interface {
	// 启动TCP服务
	Open(conf *Config, l limiter.Interface) error

	// 关闭TCP服务
	Close() error

	// 校验配置
	ValidateConfig(conf *Config) error
}

func Default() Interface {
	return New("v1")
}

func New(name string) Interface {
	if builders == nil {
		return nil
	}
	if f := builders[name]; f != nil {
		return f()
	}
	return nil
}
