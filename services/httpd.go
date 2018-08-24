package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

var ctrl = newController(1)

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
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	var start = time.Now()
	var r = &Request{
		Action:      ActionRegistQuota,
		RegistQuota: &APIRegistQuotaReq{},
	}

	statusCode, msg := resolveRequest(w, req, r.RegistQuota)
	if statusCode == 200 {
		if rp := l.Do(r); rp.Err != nil {
			statusCode = 403
			msg = rp.Err.Error()
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/registQuota]  statusCode:%d, msg:%s, elapse:%v", req.Method, statusCode, msg, time.Since(start))
}

// borrow 获取一次执行权限
func borrow(w http.ResponseWriter, req *http.Request) {
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	var start = time.Now()
	var r = &Request{
		Action: ActionBorrow,
		Borrow: &APIBorrowReq{},
	}

	statusCode, msg := resolveRequest(w, req, r.Borrow)
	if statusCode == 200 {
		if rp := l.Do(r); rp.Err != nil {
			statusCode = 403
			msg = rp.Err.Error()
		} else {
			statusCode = 200
			msg = rp.Borrow.RCID
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/borrow]  statusCode:%d, msg:%s, elapse:%v", req.Method, statusCode, msg, time.Since(start))
}

// return_ 释放一次执行权限
func return_(w http.ResponseWriter, req *http.Request) {
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	var start = time.Now()
	var r = &Request{
		Action: ActionReturn,
		Return: &APIReturnReq{},
	}

	statusCode, msg := resolveRequest(w, req, r.Return)
	if statusCode == 200 {
		if rp := l.Do(r); rp.Err != nil {
			statusCode = 403
			msg = rp.Err.Error()
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/return]  statusCode:%d, msg:%s, elapse:%v", req.Method, statusCode, msg, time.Since(start))
}

// returnAll 释放用户的所有占用的权限
func returnAll(w http.ResponseWriter, req *http.Request) {
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	var start = time.Now()
	var r = &Request{
		Action:    ActionReturnAll,
		ReturnAll: &APIReturnAllReq{},
	}

	statusCode, msg := resolveRequest(w, req, r.ReturnAll)
	if statusCode == 200 {
		if rp := l.Do(r); rp.Err != nil {
			statusCode = 403
			msg = rp.Err.Error()
		}
	}

	w.WriteHeader(statusCode)
	w.Write([]byte(msg))

	glog.V(1).Infof("%s [/v1/returnAll]  statusCode:%d, msg:%s, elapse:%v", req.Method, statusCode, msg, time.Since(start))
}

// snapshot 对元数据做快照
func snapshot(w http.ResponseWriter, req *http.Request) {
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	start := time.Now()
	defer glog.Infof("%s [/v1/snapshot] -- %v", req.Method, time.Since(start))

	// check leader
	if l.IsLeader() == false {
		addr := l.LeaderHTTPAddr()
		if len(addr) <= 0 {
			w.WriteHeader(503)
			w.Write([]byte(ErrLeaderNotFound.Error()))
			return
		}
		req.URL.Host = addr
		w.Header().Set("Location", req.URL.String())
		w.WriteHeader(307)
		return
	}

	_, err := l.Snapshot()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("Snapshot OK"))
	}
	return
}

// restore 从元数据恢复数据
func restore(w http.ResponseWriter, req *http.Request) {
	if err := ctrl.in(); err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}
	defer ctrl.out()

	start := time.Now()
	defer glog.Infof("%s [/v1/restore] -- %v", req.Method, time.Since(start))

	// check leader
	if l.IsLeader() == false {
		addr := l.LeaderHTTPAddr()
		if len(addr) <= 0 {
			w.WriteHeader(503)
			w.Write([]byte(ErrLeaderNotFound.Error()))
			return
		}
		req.URL.Host = addr
		w.Header().Set("Location", req.URL.String())
		w.WriteHeader(307)
		return
	}

	f, err := os.Open("./snapshot.limiter")
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}

	if err = l.Restore(f); err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(200)
		w.Write([]byte("Restore OK"))
	}
	return
}

func resolveRequest(w http.ResponseWriter, req *http.Request, argToResolve interface{}) (statusCode int, msg string) {
	if req.Method != http.MethodPost {
		return 405, "Method `POST` is needed"
	}

	// check leader
	if l.IsLeader() == false {
		addr := l.LeaderHTTPAddr()
		if len(addr) <= 0 {
			return 503, ErrLeaderNotFound.Error()
		}
		req.URL.Host = addr
		w.Header().Set("Location", req.URL.String())
		return 307, ""
	}

	// 读取body
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		return 500, err.Error()
	}

	// 解析参数
	if err := json.Unmarshal(b, argToResolve); err != nil {
		return 500, err.Error()
	}

	return 200, ""
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
		fmt.Printf("Enter\n")
		return nil
	default:
		return ErrTooBusy
	}
	return errors.New("Unknown error")
}

func (c *controller) out() {
	<-c.ch
}
