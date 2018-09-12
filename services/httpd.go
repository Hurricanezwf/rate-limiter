package services

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"

	_ "net/http/pprof"
)

var ctrl = newController(10000)

func init() {
	http.HandleFunc(RegistQuotaURI, registQuota)
	http.HandleFunc(DeleteQuotaURI, deleteQuota)
	http.HandleFunc(BorrowURI, borrow)
	http.HandleFunc(ReturnURI, return_)
	http.HandleFunc(ReturnAllURI, returnAll)
	http.HandleFunc(ResourceListURI, resourceList)
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
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
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
		glog.V(2).Infof("Regist '%d' quotas for resource '%s' SUCCESS", r.Quota, encoding.BytesToString(r.RCType))
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Regist '%d' quotas for resource '%s' FAILED, %s", r.Quota, encoding.BytesToString(r.RCType), msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/registQuota] RSP: statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// deleteQuota 删除资源配额
func deleteQuota(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIDeleteQuotaReq
	var rp *APIDeleteQuotaResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		goto FINISH
	}

	// 执行请求
	rp = l.DeleteQuota(&r)
	switch rp.Code {
	case 0:
		code = http.StatusOK
		glog.V(2).Infof("Delete resource '%s' SUCCESS", encoding.BytesToString(r.RCType))
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Delete resource '%s' FAILED, %s", encoding.BytesToString(r.RCType), msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/deleteQuota] RSP: statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// borrow 获取一次执行权限
func borrow(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIBorrowReq
	var rp *APIBorrowResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
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
		glog.V(2).Infof("Client '%s' borrow '%s' SUCCESS", encoding.BytesToString(r.ClientID), rp.RCID)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%s' borrow '%s' FAILED, %s", encoding.BytesToString(r.ClientID), encoding.BytesToString(r.RCType), msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/borrow] RSP: statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// return_ 释放一次执行权限
func return_(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIReturnReq
	var rp *APIReturnResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
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
		glog.V(2).Infof("Client '%s' return '%s' SUCCESS", encoding.BytesToString(r.ClientID), r.RCID)
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%s' return '%s' FAILED, %s", encoding.BytesToString(r.ClientID), r.RCID, msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/return] RSP: statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// returnAll 释放用户的所有占用的权限
func returnAll(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var r APIReturnAllReq
	var rp *APIReturnAllResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
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
		glog.V(2).Infof("Client '%s' return all resources of type '%s' SUCCESS", encoding.BytesToString(r.ClientID), encoding.BytesToString(r.RCType))
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		msg = rp.Msg
		glog.Warningf("Client '%s' return all resources of type '%s' FAILED, %s", encoding.BytesToString(r.ClientID), encoding.BytesToString(r.RCType), msg)
	}

FINISH:
	w.WriteHeader(code)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/returnAll] RSP: statusCode:%d, msg:%s, elapse:%v", req.Method, code, msg, time.Since(start))
}

// resourceList 查询资源详情列表
func resourceList(w http.ResponseWriter, req *http.Request) {
	var start = time.Now()
	var buf = bytes.NewBuffer(nil)
	var r APIResourceListReq
	var rp *APIResourceListResp

	// 解析请求
	code, msg := resolveRequest(w, req, &r, http.MethodPost)
	if code != -1 {
		defer ctrl.out()
	}
	if code != 0 {
		buf.WriteString(msg)
		goto FINISH
	}

	// 执行请求
	rp = l.ResourceList(&r)
	switch rp.Code {
	case 0:
		encoder := jsonpb.Marshaler{}
		encoder.Marshal(buf, rp)
		code = http.StatusOK
		glog.V(2).Info("List resources SUCCESS")
	case 307:
		req.URL.Host = rp.Msg
		w.Header().Set("Location", req.URL.String())
		code = http.StatusTemporaryRedirect
	default:
		code = int(rp.Code)
		buf.WriteString(rp.Msg)
		glog.Warningf("List resources FAILED, %s", buf.String())
	}

FINISH:
	w.WriteHeader(code)
	w.Write(buf.Bytes())

	glog.V(1).Infof("%s [/v1/rc] RSP: statusCode:%d, elapse:%v", req.Method, code, time.Since(start))
}

func resolveRequest(w http.ResponseWriter, req *http.Request, message proto.Message, requiredMethod string) (code int, msg string) {
	// 访问控制检测
	if err := ctrl.in(); err != nil {
		return -1, err.Error()
	}

	// 验证请求方法
	if req.Method != requiredMethod {
		return http.StatusMethodNotAllowed, fmt.Sprintf("Method `%s` is needed", requiredMethod)
	}

	// 解析请求
	buf := bytes.NewBuffer(nil)
	buf.ReadFrom(req.Body)
	req.Body.Close()

	glog.V(1).Infof("%s [%s] REQ: %s", req.Method, req.RequestURI, buf.String())

	if err := jsonpb.Unmarshal(buf, message); err != nil {
		return http.StatusBadRequest, err.Error()
	}

	return 0, ""
}

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
