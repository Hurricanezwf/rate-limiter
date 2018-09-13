package tcpserver

import (
	"bufio"
	"net"
	"time"
)

type Connection struct {
	// conn TCP连接
	conn *net.TCPConn

	// reader 连接读取器
	reader *bufio.Reader

	// keepaliveInterval 心跳间隔,单位毫秒
	keepaliveInterval int64

	// 心跳超时时间，单位毫秒
	keepaliveTimeout int64

	// 控制连接关闭
	stopC chan struct{}
}

func newConnection(conn *net.TCPConn, keepaliveInterval, keepaliveTimeout int64) *Connection {
	c := &Connection{
		conn:             conn,
		reader:           bufio.NewReader(conn),
		keepalive:        keepalive,
		keepaliveTimeout: keepaliveTimeout,
		stopC:            make(chan struct{}),
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

func (c *Connection) reader() *bufio.Reader {
	return c.reader
}

func (c *Connection) keepalive() {
	conn := c.conn
	ticker := time.NewTicker(c.keepaliveInterval * time.Millisecond)
	timer := time.NewTimer(c.keepAliveTimeout * time.Millisecond)
	for {
		select {
		case <-c.stopC:
			glog.V(3).Info("Disconnect with '%s'", conn.RemoteAddr().String())
			conn.Close()
			return
		case <-timer.C:
			glog.Warninf("Disconnect with '%s' because keepalived timeout.", conn.RemoteAddr().String())
			conn.Close()
			return
		case <-ticker.C:
			// TODO:
			// send heartbeat
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(c.keepaliveTimeout)
		}
	}
}

func (c *Connection) Close() {
	close(c.stopC)
}
