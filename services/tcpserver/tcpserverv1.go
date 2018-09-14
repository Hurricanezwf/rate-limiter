package tcpserver

import (
	"cmp/public-cloud/proxy-layer/logging/glog"
	"net"

	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/pkg/encoding"
	"github.com/Hurricanezwf/rate-limiter/proto"
)

type tcpserverv1 struct {
	// 服务器配置
	conf *Config

	// TCP监听器
	listener *net.TCPListener

	// 连接管理器
	//mgrLock *sync.RWMutex
	//mgr     map[net.Conn]struct{}

	// Limiter实例
	limiter limiter.Interface

	// dispatcher 事件转发器
	dispatcher *EventDispatcher

	// 连接数控制
	connCtrl chan struct{}

	// 控制服务器关闭
}

func newTCPServerV1() Interface {
	return &tcpserverv1{}
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

func (s *tcpserverv1) serveLoop() {
	for {
		conn, err := s.listener.AcceptTCP()
		if err != nil {
			glog.Warning("Accept error: %v", err)
			continue
		}

		// 连接数控制
		s.connCtrl <- struct{}{}

		go s.handle(newConnection(conn, s.conf.KeepAliveInterval, s.conf.KeepAliveTimeout))
	}
}

func (s *tcpserverv1) handle(c *Connection) {
	for {
		action, _, seq, msgBody, err := encoding.DecodeMsg(c.reader())
		if err != nil {
			c.Write(action, proto.TCPCodeBadRequest, seq, nil)
			goto FINISH
		}

		// 投递到队列中
		err = s.dispatcher.PushBack(&Event{
			Conn:    c,
			Seq:     seq,
			Action:  action,
			Msg:     msgBody,
			Limiter: s.l,
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
	<-s.connCtrl
}
