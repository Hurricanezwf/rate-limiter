package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
)

func init() {
	http.HandleFunc("/v1/snapshot", snapshot) // for test only
	http.HandleFunc("/v1/restore", restore)   // for test only
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
	if req.Method != http.MethodPost {
		w.WriteHeader(405)
		w.Write(ErrMsg("Method `POST` is needed"))
		return
	}

	// 读取body
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 解析参数
	var r = Request{
		Action:      ActionRegistQuota,
		RegistQuota: &APIRegistQuotaReq{},
	}
	if err := json.Unmarshal(b, r.RegistQuota); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 尝试获取资格
	if rp := l.Do(&r, true); rp.Err != nil {
		w.WriteHeader(403)
		w.Write(ErrMsg(rp.Err.Error()))
	} else {
		w.WriteHeader(200)
	}
	return
}

// borrow 获取一次执行权限
func borrow(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(405)
		w.Write(ErrMsg("Method `POST` is needed"))
		return
	}

	// 读取body
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 解析参数
	var r = Request{
		Action: ActionBorrow,
		Borrow: &APIBorrowReq{},
	}
	if err := json.Unmarshal(b, r.Borrow); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 尝试获取资格
	if rp := l.Do(&r, true); rp.Err != nil {
		w.WriteHeader(403)
		w.Write(ErrMsg(rp.Err.Error()))
	} else {
		rcId := rp.Borrow.RCID
		w.WriteHeader(200)
		w.Write([]byte(rcId))
	}
	return
}

// return_ 释放一次执行权限
func return_(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		w.WriteHeader(405)
		w.Write(ErrMsg("Method `POST` is needed"))
		return
	}

	// 读取body
	b, err := ioutil.ReadAll(req.Body)
	defer req.Body.Close()

	if err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 解析参数
	var r = Request{
		Action: ActionReturn,
		Return: &APIReturnReq{},
	}
	if err := json.Unmarshal(b, r.Return); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}

	// 尝试获取资格
	if rp := l.Do(&r, true); rp.Err != nil {
		w.WriteHeader(403)
		w.Write(ErrMsg(rp.Err.Error()))
	} else {
		w.WriteHeader(200)
	}
	return
}

// returnAll 释放用户的所有占用的权限
func returnAll(w http.ResponseWriter, req *http.Request) {
	// TODO
}

// snapshot 对元数据做快照
func snapshot(w http.ResponseWriter, req *http.Request) {
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

func ErrMsg(format string, v ...interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(format, v...))
	return buf.Bytes()
}
