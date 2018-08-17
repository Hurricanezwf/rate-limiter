package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/proto"
)

// Rate limiter实例
var l limiter.Limiter

func init() {
	http.HandleFunc("/borrow", borrow)
	http.HandleFunc("/return", return_)
	http.HandleFunc("/dead", dead)
}

// Run 启动HTTP服务
func Run(addr string) error {
	l = limiter.New()
	if err := l.Open(); err != nil {
		return fmt.Errorf("Open limiter failed, %v", err)
	}

	// 启动HTTP服务
	errC := make(chan error, 1)

	go func() {
		errC <- http.ListenAndServe(addr, nil)
	}()

	select {
	case err := <-errC:
		return fmt.Errorf("Run HTTP Server on %s failed, %v", addr, err)
	case <-time.After(3 * time.Second):
	}
	return nil
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
	if err := l.Do(&r); err != nil {
		w.WriteHeader(403)
		w.Write(ErrMsg(err.Error()))
		return
	}
	w.WriteHeader(200)
	return
}

// return_ 释放一次执行权限
func return_(w http.ResponseWriter, req *http.Request) {
	// TODO
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
