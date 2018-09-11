package ratelimiter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/golang/protobuf/jsonpb"
	uuid "github.com/satori/go.uuid"
)

var (
	ErrTooBusy             = proto.ErrTooBusy
	ErrQuotaNotEnough      = proto.ErrQuotaNotEnough
	ErrExisted             = proto.ErrExisted
	ErrResourceNotRegisted = proto.ErrResourceNotRegisted
	ErrNoBorrowRecordFound = proto.ErrNoBorrowRecordFound
	ErrQuotaOverflow       = proto.ErrQuotaOverflow
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

	// 使用过的资源类型, 这里使用base64存储
	usedRCType map[string]struct{}
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
		usedRCType:  make(map[string]struct{}),
	}

	return &c, nil
}

// Close 关闭limiter，释放该客户端占用的所有资源配额. 一般在程序退出时使用
// 注意：不调用Close操作会使得该客户端占用的资源持续占用直到超时。
func (c *RateLimiterClient) Close() error {
	var err, lastErr error
	var rcType []byte
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	for rcTypeStr := range c.usedRCType {
		if rcType, err = encoding.StringToBytes(rcTypeStr); err != nil {
			lastErr = err
			continue
		}

		buf.Reset()
		err = encoder.Marshal(buf, &proto.APIReturnAllReq{
			ClientID: c.clientId,
			RCType:   rcType,
		})
		if err != nil {
			lastErr = err
			continue
		}

		if _, err = c.sendPost("/v1/returnAll", buf); err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}

// Regist 注册资源配额
// resourceTypei : 用户自定义资源类型
// quota         : 资源配额，例如quota为10表示限流10次/s
// resetInterval : 资源配额重置周期，单位秒
func (c *RateLimiterClient) RegistQuota(resourceType []byte, quota uint32, resetInterval int64) error {
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)
	var err = encoder.Marshal(buf, &proto.APIRegistQuotaReq{
		RCType:        resourceType,
		Quota:         quota,
		ResetInterval: resetInterval,
	})

	if err == nil {
		_, err = c.sendPost("/v1/registQuota", buf)
	}

	return err
}

// Borrow 借用资源
// resourceType: 用户自定义资源类型
// expire      : 过期自动回收时间，单位秒。该时间建议与请求超时时间一致
func (c *RateLimiterClient) Borrow(resourceType []byte, expire int64) (resourceId string, err error) {
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)
	err = encoder.Marshal(buf, &proto.APIBorrowReq{
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
	c.setUsedRCType(resourceType)
	return string(rcId), nil
}

// BorrowWithTimeout 带超时的借用资源
// @resourceType: 用户自定义资源类型
// @expire      : 过期自动回收时间，单位秒。该时间建议与请求超时时间一致
func (c *RateLimiterClient) BorrowWithTimeout(resourceType []byte, expire int64, timeout time.Duration) (resourceId string, err error) {
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)
	err = encoder.Marshal(buf, &proto.APIBorrowReq{
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

	for {
		tmpBuf := bytes.NewBuffer(buf.Bytes())
		rcId, err := c.sendPost("/v1/borrow", tmpBuf)

		// 请求成功
		if err == nil {
			if len(rcId) <= 0 {
				return "", errors.New("Missing 'resourceId' in response")
			} else {
				c.setUsedRCType(resourceType)
				return string(rcId), nil
			}
		}

		// 请求失败且不是配额不足的失败，直接返回
		if err != ErrQuotaNotEnough {
			return "", err
		}

		// 对于配额不足的失败，随机sleep后重试
		randSleep = rand.Intn(900) + 100

		select {
		case <-time.After(time.Duration(randSleep) * time.Millisecond):
			// do nothing
		case <-ctx.Done():
			return "", ErrQuotaNotEnough
		}
	}
}

// Return 归还资源
// resourceType: 用户自定义资源类型
// resourceId  : 要归还的资源ID
func (c *RateLimiterClient) Return(resourceId string) error {
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)
	var err = encoder.Marshal(buf, &proto.APIReturnReq{
		ClientID: c.clientId,
		RCID:     resourceId,
	})

	if err == nil {
		_, err = c.sendPost("/v1/return", buf)
	}

	return err
}

// ResourceList 查询资源列表
// rcType : 资源类型，如果不填则表示全量查询
func (c *RateLimiterClient) ResourceList(rcType []byte) ([]*proto.APIResourceDetail, error) {
	var encoder jsonpb.Marshaler
	var buf = bytes.NewBuffer(nil)
	err := encoder.Marshal(buf, &proto.APIResourceListReq{
		RCType: rcType,
	})
	if err != nil {
		return nil, err
	}

	if body, err := c.sendPost("/v1/rc", buf); err != nil {
		return nil, err
	} else {
		buf.Reset()
		buf.Write(body)
	}

	var rp proto.APIResourceListResp
	if err = jsonpb.Unmarshal(buf, &rp); err != nil {
		return nil, err
	}

	return rp.RCList, err
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
				return nil, parseError(buf.String())
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

// setUsedRCType 记录使用过的资源类型，用户在Close的时候进行资源释放
func (c *RateLimiterClient) setUsedRCType(resourceType []byte) {
	rcTypeStr := encoding.BytesToString(resourceType)

	c.mutex.RLock()
	_, exist := c.usedRCType[rcTypeStr]
	c.mutex.RUnlock()

	if !exist {
		c.mutex.Lock()
		c.usedRCType[rcTypeStr] = struct{}{}
		c.mutex.Unlock()
	}
}

func parseError(errMsg string) error {
	switch errMsg {
	case ErrQuotaNotEnough.Error():
		return ErrQuotaNotEnough
	case ErrResourceNotRegisted.Error():
		return ErrResourceNotRegisted
	case ErrExisted.Error():
		return ErrExisted
	case ErrTooBusy.Error():
		return ErrTooBusy
	case ErrNoBorrowRecordFound.Error():
		return ErrNoBorrowRecordFound
	case ErrQuotaOverflow.Error():
		return ErrQuotaOverflow
	}
	return errors.New(errMsg)
}
