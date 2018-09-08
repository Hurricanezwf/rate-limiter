package ratelimiter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/proto"
	uuid "github.com/satori/go.uuid"
)

// 客户端配置
type ClientConfig struct {
	// 服务端集群地址
	Cluster []string
}

// RateLimiter服务客户端代理库
type RateLimiterClient struct {
	mutex *sync.RWMutex

	// 客户端ID，每个实例不重复
	clientId []byte

	// 集群地址
	cluster     []string
	clusterSize int
}

func New(config *ClientConfig) (*RateLimiterClient, error) {
	if config == nil {
		return nil, errors.New("Config is nil")
	}
	if len(config.Cluster) <= 0 {
		return nil, errors.New("Missing cluster address")
	}

	// 简单校验集群地址
	for _, addr := range config.Cluster {
		a, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, fmt.Errorf("Invalid cluster address `%s` in config", addr)
		}
		if a.IP.IsUnspecified() {
			return nil, fmt.Errorf("Invalid cluster address `%s` in config, it is not advertisable")
		}
	}

	// 生成随机ClientID
	clientId, err := uuid.NewV4()
	if err != nil {
		return nil, fmt.Errorf("Generate uuid for client failed, %v", err)
	}

	c := RateLimiterClient{
		mutex:       &sync.RWMutex{},
		clientId:    clientId.Bytes(),
		cluster:     config.Cluster,
		clusterSize: len(config.Cluster),
	}

	return &c, nil
}

// Close 关闭limiter，释放该客户端占用的所有资源配额. 一般在程序退出时使用
// 注意：不调用Close操作会使得该客户端占用的资源持续占用直到超时。
func (c *RateLimiterClient) Close() error {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(proto.APIReturnAllReq{
		ClientID: c.clientId,
	})
	if err != nil {
		return err
	}

	_, err = c.sendPost("/v1/returnAll", buf)

	return err
}

// Regist 注册资源配额
// resourceType: 用户自定义资源类型
// quota       : 资源配额，例如quota为10表示限流10次/s
func (c *RateLimiterClient) RegistQuota(resourceType []byte, quota uint32) error {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(proto.APIRegistQuotaReq{
		RCType: resourceType,
		Quota:  quota,
	})
	if err != nil {
		return err
	}

	_, err = c.sendPost("/v1/registQuota", buf)

	return err
}

// Borrow 借用资源
// resourceType: 用户自定义资源类型
// expire      : 过期自动回收时间，单位秒。该时间建议与请求超时时间一致
func (c *RateLimiterClient) Borrow(resourceType []byte, expire int64) (resourceId string, err error) {
	buf := bytes.NewBuffer(nil)
	err = json.NewEncoder(buf).Encode(proto.APIBorrowReq{
		RCType:   resourceType,
		ClientID: c.clientId,
		Expire:   expire,
	})
	if err != nil {
		return resourceId, err
	}

	rcId, err := c.sendPost("/v1/borrow", buf)
	if err != nil {
		return resourceId, err
	}
	if len(rcId) <= 0 {
		return resourceId, errors.New("Missing 'resourceId' in response")
	}
	return string(rcId), nil
}

// BorrowWithTimeout 带超时的借用资源
// resourceType: 用户自定义资源类型
// expire      : 过期自动回收时间，单位秒。该时间建议与请求超时时间一致
func (c *RateLimiterClient) BorrowWithTimeout(resourceType []byte, expire int64, timeout time.Duration) (resourceId string, err error) {
	b, err := json.Marshal(proto.APIBorrowReq{
		RCType:   resourceType,
		ClientID: c.clientId,
		Expire:   expire,
	})
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	randSleep := 0
	rand.Seed(time.Now().UnixNano())

	buf := bytes.NewBuffer(nil)

	for {
		buf.Write(b)
		rcId, err := c.sendPost("/v1/borrow", buf)

		// 请求成功
		if err == nil {
			if len(rcId) <= 0 {
				return "", errors.New("Missing 'resourceId' in response")
			} else {
				return string(rcId), nil
			}
		}

		// 请求失败且不是配额不足的失败，直接返回
		if strings.Compare(err.Error(), proto.ErrQuotaNotEnough.Error()) != 0 {
			return "", err
		}

		// 对于配额不足的失败，随机sleep后重试
		randSleep = rand.Intn(900) + 100

		select {
		case <-time.After(time.Duration(randSleep) * time.Millisecond):
			// do nothing
		case <-ctx.Done():
			return "", proto.ErrTimeout
		}
	}
}

// Return 归还资源
// resourceType: 用户自定义资源类型
// resourceId  : 要归还的资源ID
func (c *RateLimiterClient) Return(resourceId string) error {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(proto.APIReturnReq{
		ClientID: c.clientId,
		RCID:     resourceId,
	})
	if err != nil {
		return err
	}

	_, err = c.sendPost("/v1/return", buf)

	return err
}

func (c *RateLimiterClient) getCluster(idx int) string {
	if idx < 0 || idx >= c.clusterSize {
		panic("idx overflow")
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.cluster[idx]
}

func (c *RateLimiterClient) highlightLeader(leader string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for idx, addr := range c.cluster {
		if addr == leader {
			c.cluster[0], c.cluster[idx] = c.cluster[idx], c.cluster[0]
		}
	}
}

func (c *RateLimiterClient) sendPost(uri string, body io.Reader) ([]byte, error) {
	var url string
	var buf = bytes.NewBuffer(nil)

	for i := 0; i < c.clusterSize; i++ {
		host := c.getCluster(i)
		if len(url) <= 0 {
			url = fmt.Sprintf("http://%s%s", host, uri)
		}

		rp, err := http.Post(url, "application/json", body)
		if err != nil {
			if i >= c.clusterSize-1 {
				// 集群所有结点均失败
				return nil, err
			} else {
				// 部分结点失败，继续尝试
				url = ""
				continue
			}
		}

		// 请求了一个非Leader结点，需要重定向后重试
		if rp.StatusCode == 307 {
			rp.Body.Close()
			if url = rp.Header.Get("Location"); len(url) <= 0 {
				return nil, errors.New("Empty Location found in header")
			}
			continue
		}

		// 读取响应内容
		buf.Reset()
		_, err = buf.ReadFrom(rp.Body)
		rp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("Read response body failed, %v", err)
		}

		// 请求失败
		if rp.StatusCode != 200 {
			if buf.Len() > 0 {
				return nil, errors.New(buf.String())
			} else {
				return nil, fmt.Errorf("Response StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
			}
		}

		// 请求成功，检查是否需要交换集群顺序
		// 将Leader结点换到首位
		if rp.Request.URL.Host != host {
			c.highlightLeader(rp.Request.URL.Host)
		}
		break
	}
	return buf.Bytes(), nil
}
