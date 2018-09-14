package limiter

import (
	"fmt"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/cluster"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/golang/protobuf/proto"
)

func init() {
	RegistBuilder("v2", newLimiterV2)
}

// Interface limiter的接口抽象
// 负责接口验证类操作,例如参数校验，集群角色校验等。
type Interface interface {
	// Open 启用limiter
	Open() error

	// Close 关闭limiter
	Close() error

	// RegistQuota 注册资源配额
	RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp
	RegistQuotaWith(msg []byte) *APIRegistQuotaResp

	// DeleteQuota 删除资源配额
	DeleteQuota(r *APIDeleteQuotaReq) *APIDeleteQuotaResp
	DeleteQuotaWith(msg []byte) *APIDeleteQuotaResp

	// Borrow 借资源
	Borrow(r *APIBorrowReq) *APIBorrowResp
	BorrowWith(msg []byte) *APIBorrowResp

	// Return 归还单个资源
	Return(r *APIReturnReq) *APIReturnResp
	ReturnWith(msg []byte) *APIReturnResp

	// ReturnAll 归还用户借取的所有资源，通常用于客户端关闭时
	ReturnAll(r *APIReturnAllReq) *APIReturnAllResp
	ReturnAllWith(msg []byte) *APIReturnAllResp

	// ResourceList 查询资源列表详情
	ResourceList(r *APIResourceListReq) *APIResourceListResp
	ResourceListWith(msg []byte) *APIResourceListResp

	// IsOpen 是否已经开启
	IsOpen() bool
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
	gLock *sync.RWMutex

	isOpen bool

	c cluster.Interface
}

func newLimiterV2() Interface {
	return &limiterV2{
		gLock:  &sync.RWMutex{},
		isOpen: false,
		c:      cluster.Default(),
	}
}

func (l *limiterV2) Open() error {
	err := l.c.Open()
	l.gLock.Lock()
	l.isOpen = (err == nil)
	l.gLock.Unlock()
	return err
}

func (l *limiterV2) Close() error {
	l.gLock.Lock()
	l.isOpen = false
	l.gLock.Unlock()
	return l.c.Close()
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

func (l *limiterV2) RegistQuotaWith(msg []byte) *APIRegistQuotaResp {
	var r APIRegistQuotaReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIRegistQuotaResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.RegistQuota(&r)
}

func (l *limiterV2) DeleteQuota(r *APIDeleteQuotaReq) *APIDeleteQuotaResp {
	var rp APIDeleteQuotaResp

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

	return l.c.DeleteQuota(r)
}

func (l *limiterV2) DeleteQuotaWith(msg []byte) *APIDeleteQuotaResp {
	var r APIDeleteQuotaReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIDeleteQuotaResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.DeleteQuota(&r)
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

func (l *limiterV2) BorrowWith(msg []byte) *APIBorrowResp {
	var r APIBorrowReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIBorrowResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.Borrow(&r)
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

func (l *limiterV2) ReturnWith(msg []byte) *APIReturnResp {
	var r APIReturnReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIReturnResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.Return(&r)
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

func (l *limiterV2) ReturnAllWith(msg []byte) *APIReturnAllResp {
	var r APIReturnAllReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIReturnAllResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.ReturnAll(&r)
}

func (l *limiterV2) ResourceList(r *APIResourceListReq) *APIResourceListResp {
	var rp APIResourceListResp

	if l.c.IsLeader() == false {
		rp.Code = 307
		rp.Msg = l.c.LeaderHTTPAddr()
		return &rp
	}

	return l.c.ResourceList(r)
}

func (l *limiterV2) ResourceListWith(msg []byte) *APIResourceListResp {
	var r APIResourceListReq
	if err := proto.Unmarshal(msg, &r); err != nil {
		return &APIResourceListResp{
			Code: 400,
			Msg:  fmt.Sprintf("Bad request, %v", err),
		}
	}
	return l.ResourceList(&r)
}

func (l *limiterV2) IsOpen() bool {
	l.gLock.RLock()
	defer l.gLock.RUnlock()
	return l.isOpen
}
