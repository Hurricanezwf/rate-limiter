package tcp

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/Hurricanezwf/rate-limiter/pkg/encoding"
	"github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

type ServerConfig struct {
	// TCP服务器监听地址, 非0.0.0.0
	Listen string

	// 最大连接数
	MaxConnection int

	// 连接配置
	*ConnectionConfig

	// 转发器配置
	*DispatcherConfig
}

type Server interface {
	// 启动TCP服务
	Open(conf *ServerConfig, router EventRouter) error

	// 关闭TCP服务
	Close() error

	// 校验配置
	ValidateConfig(conf *ServerConfig) error
}

func Default() Server {
	return &tcpserverv1{}
}

type tcpserverv1 struct {
	// 服务器配置
	conf *ServerConfig

	// TCP监听器
	listener *net.TCPListener

	// dispatcher 事件转发器
	dispatcher *EventDispatcher

	// 连接数控制
	maxConnCtrl chan struct{}

	// 控制服务器关闭
	stopC chan struct{}
}

// Open 打开TCP服务器
func (s *tcpserverv1) Open(conf *ServerConfig, router EventRouter) error {
	// 校验配置
	if err := s.ValidateConfig(conf); err != nil {
		return fmt.Errorf("Config error: %v", err)
	}

	// 开启事件分发
	dispatcher := &EventDispatcher{}
	if err := dispatcher.Open(conf.DispatcherConfig, router); err != nil {
		return err
	}

	// 启动监听
	addr, err := net.ResolveTCPAddr("tcp", conf.Listen)
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	// 变量初始化
	s.conf = conf
	s.listener = listener
	s.dispatcher = dispatcher
	s.maxConnCtrl = make(chan struct{}, conf.MaxConnection)
	s.stopC = make(chan struct{})

	// 启动服务
	go s.serveLoop()

	return nil
}

func (s *tcpserverv1) serveLoop() {
	var dropCount uint64 = 0

	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			glog.Warning("Accept error: %v", err)
			continue
		}

		// 连接数控制, 3秒超时
		select {
		case s.maxConnCtrl <- struct{}{}:
			// do nothing
		case <-time.After(3 * time.Second):
			conn.Close()
			dropCount++
			if dropCount%uint64(100) == uint64(1) {
				glog.Warningf("Drop connection because I'm too busy. DropCount=%d", dropCount)
			}
		}

		go s.handle(newConnection(conn, s.conf.ConnectionConfig))
	}
}

func (s *tcpserverv1) handle(c *Connection) {
	for {
		// 检测服务器关闭
		if s.IsOpen() == false {
			goto FINISH
		}

		// 读取请求
		action, _, seq, msgBody, err := encoding.DecodeMsg(c.Reader())
		if err != nil {
			c.Write(action, proto.TCPCodeBadRequest, seq, nil)
			goto FINISH
		}

		// 投递到队列中
		err = s.dispatcher.PushBack(&Event{
			Conn:   c,
			Seq:    seq,
			Action: action,
			Msg:    msgBody,
		})
		if err == nil {
			continue
		}
		if err = c.Write(action, proto.TCPCodeServerTooBusy, seq, nil); err != nil {
			glog.Warningf("Write to remote '%s' failed, ", c.RemoteAddr(), err.Error())
			continue
		}
	}
FINISH:
	c.Close()
	<-s.maxConnCtrl
}

func (s *tcpserverv1) IsOpen() bool {
	select {
	case <-s.stopC:
		return false
	default:
		return true
	}
}

func (s *tcpserverv1) ValidateConfig(conf *ServerConfig) error {
	if conf == nil {
		return errors.New("Config is nil")
	}

	//
	addr, err := net.ResolveTCPAddr("tcp", conf.Listen)
	if err != nil {
		return fmt.Errorf("Invalid `Listen` field value, %v", err)
	}
	if addr.IP.IsUnspecified() {
		return errors.New("Invalid `Listen` field value, it is not advertisable")
	}

	//
	if conf.MaxConnection <= 0 {
		return fmt.Errorf("Invalid `MaxConnection` field value")
	}

	//
	if err = ValidateConnectionConf(conf.ConnectionConfig); err != nil {
		return err
	}

	//
	if err = ValidateDispatcherConf(conf.DispatcherConfig); err != nil {
		return err
	}
	return nil
}

func (s *tcpserverv1) Close() error {
	select {
	case <-s.stopC:
		return nil
	default:
		s.listener.Close()
		s.dispatcher.Close()
		close(s.stopC)
		close(s.maxConnCtrl)
	}
	return nil
}