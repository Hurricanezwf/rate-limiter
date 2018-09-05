package services

import (
	"bytes"
	"errors"
	"net/http"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var ctrl = newController(10000)

func init() {
	//http.HandleFunc("/v1/snapshot", snapshot) // for test only
	//http.HandleFunc("/v1/restore", restore)   // for test only
	http.HandleFunc("/v1/registQuota", registQuota)
	http.HandleFunc("/v1/borrow", borrow)
	http.HandleFunc("/v1/return", return_)
	http.HandleFunc("/v1/returnAll", returnAll)
}

// runHttpd 启动HTTP服务
func runHttpd(addr string) error {
	// 启动HTTP服务
	errC := make(chan error, 1)

	go func() {
		errC <- http.ListenAndServe(addr, nil)
	}()

	select {
	case err := <-errC:
		return err
	case <-time.After(time.Second):
	}
	return nil
}

// registQuota 注册资源配额
func registQuota(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIRegistQuotaReq
	var rp *APIRegistQuotaResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		goto FINISH
	}

	// 执行请求
	rp = l.RegistQuota(&r)
	switch rp.Code {
	case 0:
		code = http.StatusOK
		glog.V(2).Infof("Regist '%d' quotas for resource '%x' SUCCESS", r.Quota, r.RCType)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Regist '%d' quotas for resource '%x' FAILED, %s", r.Quota, r.RCType, msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/registQuota]  statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// borrow 获取一次执行权限
func borrow(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIBorrowReq
	var rp *APIBorrowResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		goto FINISH
	}

	// 执行请求
	rp = l.Borrow(&r)
	switch rp.Code {
	case 0:
		code = http.StatusOK
		msg = rp.RCID
		glog.V(2).Infof("Client '%x' borrow '%s' SUCCESS", r.ClientID, rp.RCID)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%x' borrow '%x' FAILED, %s", r.ClientID, r.RCType, msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/borrow]  statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// return_ 释放一次执行权限
func return_(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIReturnReq
	var rp *APIReturnResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		goto FINISH
	}

	// 执行请求
	rp = l.Return(&r)
	switch rp.Code {
	case 0:
		code = http.StatusOK
		glog.V(2).Infof("Client '%x' return '%s' SUCCESS", r.ClientID, r.RCID)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%x' return '%s' FAILED, %s", r.ClientID, r.RCID, msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/return]  statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// returnAll 释放用户的所有占用的权限
func returnAll(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIReturnAllReq
	var rp *APIReturnAllResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		goto FINISH
	}

	// 执行请求
	rp = l.ReturnAll(&r)
	switch rp.Code {
	case 0:
		code = http.StatusOK
		glog.V(2).Infof("Client '%x' return all resources of type '%x' SUCCESS", r.ClientID, r.RCType)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%x' return all resources of type '%x' FAILED, %s", r.ClientID, r.RCType, msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/returnAll]  statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// snapshot 对元数据做快照
//func snapshot(w http.ResponseWriter, req *http.Request) {
//	if err := ctrl.in(); err != nil {
//		w.WriteHeader(403)
//		w.Write([]byte(err.Error()))
//		return
//	}
//	defer ctrl.out()
//
//	start := time.Now()
//	defer glog.Infof("%s [/v1/snapshot] -- %v", req.Method, time.Since(start))
//
//	// check leader
//	if l.IsLeader() == false {
//		addr := l.LeaderHTTPAddr()
//		if len(addr) <= 0 {
//			w.WriteHeader(503)
//			w.Write([]byte(ErrLeaderNotFound.Error()))
//			return
//		}
//		req.URL.Host = addr
//		w.Header().Set("Location", req.URL.String())
//		w.WriteHeader(307)
//		return
//	}
//
//	_, err := l.Snapshot()
//	if err != nil {
//		w.WriteHeader(500)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(200)
//		w.Write([]byte("Snapshot OK"))
//	}
//	return
//}

// restore 从元数据恢复数据
//func restore(w http.ResponseWriter, req *http.Request) {
//	if err := ctrl.in(); err != nil {
//		w.WriteHeader(403)
//		w.Write([]byte(err.Error()))
//		return
//	}
//	defer ctrl.out()
//
//	start := time.Now()
//	defer glog.Infof("%s [/v1/restore] -- %v", req.Method, time.Since(start))
//
//	// check leader
//	if l.IsLeader() == false {
//		addr := l.LeaderHTTPAddr()
//		if len(addr) <= 0 {
//			w.WriteHeader(503)
//			w.Write([]byte(ErrLeaderNotFound.Error()))
//			return
//		}
//		req.URL.Host = addr
//		w.Header().Set("Location", req.URL.String())
//		w.WriteHeader(307)
//		return
//	}
//
//	f, err := os.Open("./snapshot.limiter")
//	if err != nil {
//		w.WriteHeader(500)
//		w.Write([]byte(err.Error()))
//		return
//	}
//
//	if err = l.Restore(f); err != nil {
//		w.WriteHeader(500)
//		w.Write([]byte(err.Error()))
//	} else {
//		w.WriteHeader(200)
//		w.Write([]byte("Restore OK"))
//	}
//	return
//}

func resolveRequest(w http.ResponseWriter, req *http.Request, message proto.Message) (code int, msg string) {
	// 访问控制检测
	if err := ctrl.in(); err != nil {
		return -1, err.Error()
	}

	// 验证请求方法
	if req.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, "Method `POST` is needed"
	}

	// 解析请求
	buf := bytes.NewBuffer(make([]byte, 0, 1024))
	buf.ReadFrom(req.Body)
	req.Body.Close()

	glog.V(4).Infof("%s [%s] %s", req.Method, req.RequestURI, buf.String())

	if err := jsonpb.Unmarshal(buf, message); err != nil {
		return http.StatusBadRequest, err.Error()
	}

	return 0, ""
}

//func resolveRequest(w http.ResponseWriter, req *http.Request, argToResolve interface{}) (statusCode int, msg string) {
//	if req.Method != http.MethodPost {
//		return 405, "Method `POST` is needed"
//	}
//
//	// check leader
//	if l.IsLeader() == false {
//		addr := l.LeaderHTTPAddr()
//		if len(addr) <= 0 {
//			return 503, ErrLeaderNotFound.Error()
//		}
//		req.URL.Host = addr
//		w.Header().Set("Location", req.URL.String())
//		return 307, ""
//	}
//
//	// 读取body
//	b, err := ioutil.ReadAll(req.Body)
//	defer req.Body.Close()
//
//	if err != nil {
//		return 500, err.Error()
//	}
//
//	// 解析参数
//	if err := json.Unmarshal(b, argToResolve); err != nil {
//		return 500, err.Error()
//	}
//
//	return 200, ""
//}

// controller 访问控制器
type controller struct {
	limit int

	ch chan struct{}
}

func newController(n int) *controller {
	return &controller{
		limit: n,
		ch:    make(chan struct{}, n),
	}
}

func (c *controller) in() error {
	select {
	case c.ch <- struct{}{}:
		return nil
	default:
		return ErrTooBusy
	}
	return errors.New("Unknown error")
}

func (c *controller) out() {
	<-c.ch
}
