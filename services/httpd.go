package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Hurricanezwf/rate-limiter/proto"
)

func init() {
	http.HandleFunc("/v1/registQuota", registQuota)
	http.HandleFunc("/v1/borrow", borrow)
	http.HandleFunc("/v1/return", return_)
	http.HandleFunc("/v1/dead", dead)
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
	case <-time.After(3 * time.Second):
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
	var r proto.Request
	if err := json.Unmarshal(b, &r); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	} else {
		r.Action = proto.ActionRegistQuota
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
	var r proto.Request
	if err := json.Unmarshal(b, &r); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	} else {
		r.Action = proto.ActionBorrow
	}

	// 尝试获取资格
	if rp := l.Do(&r, true); rp.Err != nil {
		w.WriteHeader(403)
		w.Write(ErrMsg(rp.Err.Error()))
	} else {
		rcId := rp.Result.(string)
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
	var r proto.Request
	if err := json.Unmarshal(b, &r); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	} else {
		r.Action = proto.ActionReturn
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

// dead 释放用户的所有占用的权限
func dead(w http.ResponseWriter, req *http.Request) {
	// TODO
}

func ErrMsg(format string, v ...interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf(format, v...))
	return buf.Bytes()
}
