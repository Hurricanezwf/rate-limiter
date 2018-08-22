package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hurricanezwf/rate-limiter/services"
	_ "github.com/Hurricanezwf/toolbox/logging"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

func main() {
	// 解析命令行参数
	flag.Parse()
	defer glog.Flush()

	// 显示版本信息
	ShowVersion()

	// 启动所有服务
	if err := services.Run(); err != nil {
		glog.Fatalf(err.Error())
	}

	// 捕获信号等待终止
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	glog.Warningf("Receive %s, so exit", s)
}

// 编译信息
var (
	CompileTime string
	GoVersion   string
	Branch      string
	Commit      string
)

func ShowVersion() {
	glog.Infof("CompileTime : %s", CompileTime)
	glog.Infof("GoVersion   : %s", GoVersion)
	glog.Infof("Branch      : %s", Branch)
	glog.Infof("Commit      : %s", Commit)
	glog.Info()
}
