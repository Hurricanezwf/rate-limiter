package ratelimiter

import (
	"errors"
	"fmt"
	"time"

	"github.com/Hurricanezwf/rate-limiter/proto"
)

func init() {
	RegistBuilder("http", NewHTTPClient)
}

var (
	ErrTooBusy             = proto.ErrTooBusy
	ErrQuotaNotEnough      = proto.ErrQuotaNotEnough
	ErrExisted             = proto.ErrExisted
	ErrResourceNotRegisted = proto.ErrResourceNotRegisted
	ErrNoBorrowRecordFound = proto.ErrNoBorrowRecordFound
	ErrQuotaOverflow       = proto.ErrQuotaOverflow
)

// Interface 接口抽象
type Interface interface {
	// Close 关闭客户端
	// 该操作通常在退出应用进程之前执行，用于将该客户端占用的资源配额主动释放掉
	Close() error

	// RegistQuota 注册资源配额
	// 如果资源类型已经存在，将返回ErrExisted
	//
	// @resourceType	: 资源类型
	// @quota			: 配额值
	// @resetInterval	: 资源配额重置的时间间隔, 单位秒
	RegistQuota(resourceType []byte, quota uint32, resetInterval int64) error

	// DeleteQuota 删除资源配额
	// 如果资源未找到，将返回ErrResourceNotRegisted
	//
	// @resourceType	: 资源类型
	DeleteQuota(resourceType []byte) error

	// Borrow 借用资源
	// 调用该操作时，会从服务器申请一个资源并返回资源ID, 该函数只会请求服务器一次。
	// 如果资源未注册，将返回ErrResourceNotRegisted.
	// 如果资源配额不足将返回ErrQuotaNotEnough.
	// 如果在过期时间内没有将资源归还，其将被自动回收，建议设置的expire时间小于等于请求超时时间
	//
	// @resourceType	: 资源类型
	// @exipre			: 资源自动过期时间，单位秒
	// @resourceId		: 返回申请的资源ID
	Borrow(resourceType []byte, expire int64) (resourceId string, err error)

	// BorrowWithTimeout 带超时的借用资源
	// 调用该操作时，会从服务器申请一个资源并返回资源ID，该函数会在指定的超时时间内持续请求服务器直到成功或者超时。
	//
	// 如果资源未注册，将返回ErrResourceNotRegisted.
	// 如果资源配额不足将返回ErrQuotaNotEnough.
	// 如果在超时时间内均未申请到资源，将返回ErrQuotaNotEnough
	// 如果在过期时间内没有将资源归还，其将被自动回收，建议设置的expire时间小于等于请求超时时间
	//
	// @resourceType	: 资源类型
	// @exipre			: 资源自动过期时间，单位秒
	// @resourceId		: 返回申请的资源ID
	BorrowWithTimeout(resourceType []byte, expire int64, timeout time.Duration) (resourceId string, err error)

	// Return 归还申请的资源
	// 一般用于单个资源的归还
	//
	// @resourceId		: 申请的资源ID
	Return(resourceId string) error
}

// 客户端配置
type ClientConfig struct {
	// 服务端集群地址
	Cluster []string
}

func New(name string, conf *ClientConfig) (Interface, error) {
	if builders == nil {
		return nil, errors.New("Builder is nil")
	}
	if f := builders[name]; f != nil {
		return f(conf)
	}
	return nil, fmt.Errorf("No ratelimiter client found for '%s'", name)
}
