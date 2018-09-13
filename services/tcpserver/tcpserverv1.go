package tcpserver

import (
	"cmp/public-cloud/proxy-layer/logging/glog"
	"encoding/binary"
	"io"
	"net"

	"github.com/Hurricanezwf/rate-limiter/event"
	"github.com/Hurricanezwf/rate-limiter/limiter"
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

// 协议格式:
// | Magic Number | Data Len | Data Bytes |
// |     1Byte    |   2Bytes |   NBytes   |
func (s *tcpserverv1) handle(c *Connection) {
	var (
		err         error
		b           int
		magicNumber byte
		action      byte
		dataLen     uint32
		reader      = c.reader()
	)

	for {
		// 魔数
		magicNumber, err = reader.ReadByte()
		if err != nil {
			glog.Warning(err.Error())
			break
		}
		if magicNumber != proto.MagicNumber {
			glog.Warning("MagicNumer %#v not match", magicNumber)
			break
		}

		// Action
		action, err = reader.ReadByte()
		if err != nil {
			glog.Warning(err.Error())
			break
		}

		// 数据长度
		err = binary.Read(reader, binary.BigEndian, &dataLen)
		if err != nil {
			glog.Warning(err.Error())
			break
		}

		// 数据内容, 这里msgBytes务必重新分配空间，因为它会被作为参数传递进Event
		msgBytes := make([]byte, dataLen)
		_, err = reader.Read(msgByte)
		if err != nil && err != io.EOF {
			glog.Warning(err.Error())
			break
		}

		// 投递到队列中
		err = event.PushBack(&event.Event{
			Action:  action,
			Msg:     msgByte,
			Conn:    c,
			Limiter: s.l,
		})
		// TODO:
	}
FINISH:
	c.Close()
	<-s.connCtrl
}
