package limiter

import (
	"github.com/Hurricanezwf/rate-limiter/cluster"
	. "github.com/Hurricanezwf/rate-limiter/proto"
)

func init() {
	RegistBuilder("v2", newLimiterV2)
}

// Interface limiter的接口抽象
// 负责接口验证类操作,例如参数校验，集群角色校验等。
type Interface interface {
	// Open 启用limiter
	Open() error

	// RegistQuota 注册资源配额
	RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp

	// Borrow 借资源
	Borrow(r *APIBorrowReq) *APIBorrowResp

	// Return 归还单个资源
	Return(r *APIReturnReq) *APIReturnResp

	// ReturnAll 归还用户借取的所有资源，通常用于客户端关闭时
	ReturnAll(r *APIReturnAllReq) *APIReturnAllResp
}

// Default 新建一个默认limiter实例
func Default() Interface {
	return New("v2")
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
	c cluster.Interface
}

func newLimiterV2() Interface {
	return &limiterV2{
		c: cluster.Default(),
	}
}

func (l *limiterV2) Open() error {
	return l.c.Open()
}

func (l *limiterV2) RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp {
	var rp APIRegistQuotaResp

	if l.c.IsLeader() == false {
		rp.Code = 307
		rp.Msg = l.c.LeaderHTTPAddr()
		return &rp
	}
	if len(r.RCType) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'RCType' field"
		return &rp
	}
	if r.Quota <= 0 || r.Quota > 1000000 {
		rp.Code = 403
		rp.Msg = "Invalid 'Quota' value range, it should be in (0, 100000]"
		return &rp
	}
	if r.ResetInterval <= 0 {
		rp.Code = 403
		rp.Msg = "Invalid 'ResetInterval' value, it should be greater then zero"
		return &rp
	}

	return l.c.RegistQuota(r)
}

func (l *limiterV2) Borrow(r *APIBorrowReq) *APIBorrowResp {
	var rp APIBorrowResp

	if l.c.IsLeader() == false {
		rp.Code = 307
		rp.Msg = l.c.LeaderHTTPAddr()
		return &rp
	}
	if len(r.RCType) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'RCType' field"
		return &rp
	}
	if len(r.ClientID) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'ClientID' field"
		return &rp
	}
	if r.Expire <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'Expire' field"
		return &rp
	}

	return l.c.Borrow(r)
}

func (l *limiterV2) Return(r *APIReturnReq) *APIReturnResp {
	var rp APIReturnResp

	if l.c.IsLeader() == false {
		rp.Code = 307
		rp.Msg = l.c.LeaderHTTPAddr()
		return &rp
	}
	if len(r.RCID) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'RCID' field"
		return &rp
	}
	if len(r.ClientID) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'ClientID' field"
		return &rp
	}

	return l.c.Return(r)
}

func (l *limiterV2) ReturnAll(r *APIReturnAllReq) *APIReturnAllResp {
	var rp APIReturnAllResp

	if l.c.IsLeader() == false {
		rp.Code = 307
		rp.Msg = l.c.LeaderHTTPAddr()
		return &rp
	}
	if len(r.RCType) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'RCType' field"
		return &rp
	}
	if len(r.ClientID) <= 0 {
		rp.Code = 403
		rp.Msg = "Missing 'ClientID' field"
		return &rp
	}

	return l.c.ReturnAll(r)
}
