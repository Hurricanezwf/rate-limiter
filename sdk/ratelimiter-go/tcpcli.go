package ratelimiter

import (
	"bufio"
	"cmp/public-cloud/proxy-layer/logging/glog"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/pkg/encoding"
	"github.com/Hurricanezwf/rate-limiter/proto"
	uuid "github.com/satori/go.uuid"
)

// RateLimiter客户端代理库，走TCP协议
type TCPClient struct {

	// 配置
	config *ClientConfig

	// 客户端ID，每个实例不重复
	clientId []byte

	// TCP Client
	tcpClient *net.TCPConn

	// 序列号
	seq uint32

	// 控制客户端关闭
	stopC chan struct{}

	mutex *sync.RWMutex
	// 使用过的资源类型, 这里使用base64存储
	usedRCType map[string]struct{}
	// 连接记录
	connections []*net.TCPConn
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

	c := &TCPClient{
		clientId:   clientId.Bytes(),
		config:     config,
		tcpClient:  *net.TCPAddr,
		seq:        0,
		stopC:      make(chan struct{}),
		mutex:      &sync.RWMutex{},
		usedRCType: make(map[string]struct{}),
	}
	if err = c.dial(); err != nil {
		return err
	}

	return c
}

func (c *TCPClient) dial() error {
	for _, server := range c.config.Cluster {
		addr, err := net.ResolveTCPAddr("tcp", server)
		if err != nil {
			return err
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			glog.Warningf("Dial %s failed, %v", server, err)
			continue
		}
		ok = true
		c.connections = append(c.connections, conn)
	}

	if len(c.connections) <= 0 {
		return errors.New("Dial cluster failed.")
	}
	return nil
}

func (c *TCPClient) RegistQuota(resourceType []byte, quota uint32, resetInterval int64) error {
	req := &proto.APIRegistQuotaReq{
		RCType:        resourceType,
		Quota:         quota,
		ResetInterval: resetInterval,
	}

	msg, err := encoding.EncodeMsg(proto.ActionRegistQuota, 0, c.genSeq(), req)
	if err == nil {
		_, err = c.sendRequest(msg)
	}

	return err
}

func (c *TCPClient) genSeq() uint32 {
	if c.seq == (1<<32 - 1) {
		c.seq = 0
	} else {
		c.seq++
	}
	return c.seq
}

func (c *TCPClient) sendRequest(msg []byte) <-chan *response {
	var lastErr error
	var response []byte

	for _, conn := range c.connections {
		if err := conn.Write(msg); err != nil {
			lastErr = err
			glog.Warning(err.Error())
			continue
		}

		buf := bufio.NewReader(conn)
		action, resultCode, seq, msgBody, err := encoding.DecodeMsg(buf)
		if err != nil {
			lastErr = err
			glog.Warning(err.Error())
			continue
		}
		if resultCode != proto.TCPCodeOK {
			lastErr = fmt.Errorf("Return code(%d) != %d, %s", string(msgBody))
			glog.Warning(err.Error())
			continue
		}
	}
}

type response struct {
	Err  error
	Body []byte
}

//type connection struct {
//	conn *net.TCPConn
//
//	addr *net.TCPAddr
//
//	gLock     *sync.RWMutex
//	available bool
//}
