package limiter

import (
	. "github.com/Hurricanezwf/rate-limiter/proto"
)

// Interface limiter的接口抽象
type Interface interface {
	// Open 启用limiter
	Open() error

	// Close 关闭limiter
	Close() error

	// Do handle request
	Do(r *Request) *Response
}

//func New(name string) Limiter {
//	if limiterBuilders == nil {
//		return nil
//	}
//	if f := limiterBuilders[name]; f != nil {
//		return f()
//	}
//	return nil
//}
//
//var limiterBuilders = make(map[string]limiterBuilder)
//
//type limiterBuilder func() Limiter
//
//func RegistLimiterBuilder(name string, f limiterBuilder) {
//	if f == nil {
//		panic(fmt.Sprintf("LimiterBuilder to regist is nil, name is %s", name))
//	}
//	if _, exist := limiterBuilders[name]; exist {
//		panic(fmt.Sprintf("LimiterBuilder for '%s' had been existed", name))
//	}
//	limiterBuilders[name] = f
//}
