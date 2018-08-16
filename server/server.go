package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func init() {
	http.HandleFunc("/get", get)
	http.HandleFunc("/put", put)
	http.HandleFunc("/close", putAll)
}

func Run(addr string) error {
	return http.ListenAndServe(addr, nil)
}

// get 获取一次执行权限
func get(w http.ResponseWriter, req *http.Request) {
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
	var r APIGetReq
	if err := json.Unmarshal(b, &r); err != nil {
		w.WriteHeader(500)
		w.Write(ErrMsg(err.Error()))
		return
	}
}

// put 释放一次执行权限
func put(w http.ResponseWriter, req *http.Request) {

}

// close 释放用户的所有占用的权限
func putAll(w http.ResponseWriter, req *http.Request) {

}

func ErrMsg(format string, v ...interface{}) []byte {
	return []byte{fmt.Sprintf(format, v)}
}
