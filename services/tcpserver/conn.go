package tcpserver

import (
	"bufio"
	"fmt"
	"net"
	"time"

	"github.com/Hurricanezwf/rate-limiter/pkg/encoding"
	"github.com/golang/protobuf/proto"
)

type ConnectionConfig struct {
	// 心跳间隔
	KeepAliveInterval time.Duration

	// 心跳超时时间
	KeepAliveTimeout time.Duration

	// 读写超时
	RWTimeout time.Duration
}

type Connection struct {
	// 配置
	conf *ConnectionConfig

	// conn TCP连接
	conn *net.TCPConn

	// reader 连接读取器
	reader *bufio.Reader

	// 控制连接关闭
	stopC chan struct{}
}

func newConnection(conn *net.TCPConn, conf *ConnectionConfig) *Connection {
	conn.SetNoDelay(true)
	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(conf.KeepAliveInterval)

	c := &Connection{
		conf:   conf,
		conn:   conn,
		reader: bufio.NewReader(conn),
		stopC:  make(chan struct{}),
	}

	go c.keepalive()

	return c
}

func (c *Connection) IsOpen() bool {
	select {
	case <-c.stopC:
		return false
	}
	return true
}

func (c *Connection) RemoteAddr() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) Write(action, resultCode byte, seq uint32, msg proto.Message) error {
	b, err := encoding.EncodeMsg(action, resultCode, seq, msg)
	if err != nil {
		return fmt.Errorf("Encode msg failed, %v", err)
	}

	timeoutAt := time.Now().Add(c.conf.RWTimeout)
	c.conn.SetWriteDeadline(timeoutAt)
	_, err = c.conn.Write(b)
	return err
}

func (c *Connection) reader() *bufio.Reader {
	return c.reader
}

func (c *Connection) keepalive() {
	conn := c.conn
	ticker := time.NewTicker(c.conf.KeepAliveInterval)
	timer := time.NewTimer(c.conf.KeepAliveTimeout)
	for {
		select {
		case <-c.stopC:
			glog.V(3).Info("Disconnect with '%s'", c.RemoteAddr())
			conn.Close()
			return
		case <-timer.C:
			glog.Warninf("Disconnect with '%s' because keepalived timeout.", c.RemoteAddr())
			conn.Close()
			return
		case <-ticker.C:
			// TODO:
			// send heartbeat
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(c.KeepAliveTimeout)
		}
	}
}

func (c *Connection) Close() {
	close(c.stopC)
}
