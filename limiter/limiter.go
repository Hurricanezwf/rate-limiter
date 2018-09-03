package limiter

import (
	"github.com/Hurricanezwf/rate-limiter/limiter"
	. "github.com/Hurricanezwf/rate-limiter/proto"
)

func init() {
	RegistBuilder("v2", newLimiterV2)
}

// Interface limiter的接口抽象
type Interface interface {
	// Open 启用limiter
	Open() error

	// RegistQuota 注册资源配额
	RegistQuota(req *APIRegistQuotaReq) *APIRegistQuotaResp

	// Borrow 借资源
	Borrow(req *APIBorrowReq) *APIBorrowResp

	// Return 归还单个资源
	Return(req *APIReturnReq) *APIReturnResp

	// ReturnAll 归还用户借取的所有资源，通常用于客户端关闭时
	ReturnAll(req *APIReturnAllReq) *APIReturnAllResp
}

// New 新建一个limiter接口实例
func New(name string) Interface {
	if builders == nil {
		return nil
	}
	if f := builders[name]; f != nil {
		return f()
	}
	return nil
}

// limiterV2 limiter的具体实现
type limiterV2 struct {
	cluster.Interface
}

func newLimiterV2() limiter.Interface {
	return &limiterV2{
		cluster.New("v2"),
	}
}

func (l *limiterV2) Open() error {
	return l.c.Open()
}

func (l *limiterV2) Do(r *Request) *Response {
	return nil
}
