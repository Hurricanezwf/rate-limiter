package app

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/services"
	"github.com/Hurricanezwf/rate-limiter/version"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

const name = "rate-limiter"

func init() {
	AddStartHook(g.Init)
	AddStartHook(version.Display)
	AddStartHook(services.Run)
}

func Run() {
	flag.Parse()
	defer glog.Flush()

	// 启动服务钩子函数
	for _, f := range startHooks {
		if err := f(); err != nil {
			glog.Fatal(err.Error())
		}
	}

	glog.V(1).Infof("Start %s SUCCESS", name)

	// 捕获信号
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	s := <-sig
	glog.Warningf("Receive signal %v, exit now", s)

	// 设定关闭超时后强制退出
	go forceExitWhenTimeout(30 * time.Second)

	// 执行服务停止钩子函数
	for _, f := range stopHooks {
		f()
	}
}

//func initHTTPServer() error {
//	go beego.Run(fmt.Sprintf(":%d", g.HTTPort()))
//	glog.V(1).Infof("Init %16s ................................ [OK]", "HTTPServer")
//	return nil
//}

func forceExitWhenTimeout(timeout time.Duration) {
	<-time.After(timeout)
	os.Exit(-1)
}
