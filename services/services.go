package services

import (
	"fmt"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/services/tcpserver"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

// Rate limiter实例
var l limiter.Interface

var tcpserv tcpserver.Interface

func Run() (err error) {
	glog.Infof("Be starting, wait a while...")

	// 启动HTTP服务
	if err := runHttpd(g.Config.Httpd.Listen); err != nil {
		return fmt.Errorf("Run HTTP Server on %s failed, %v", g.Config.Httpd.Listen, err)
	}
	glog.Infof("Run HTTP Service OK, listen at %s.", g.Config.Httpd.Listen)

	// limiter依赖于httpd服务
	l = limiter.Default()
	if err = l.Open(); err != nil {
		return fmt.Errorf("Open limiter failed, %v", err)
	}
	glog.Info("Open limiter OK.")

	// 启动TCP服务
	tcpserv = tcpserver.Default()
	err = tcpserv.Open(
		&tcpserver.Config{
			Listen:        "127.0.0.1:19999",
			MaxConnection: 100,
			ConnectionConfig: &tcpserver.ConnectionConfig{
				KeepAliveInterval: time.Second,
				KeepAliveTimeout:  3 * time.Second,
				RWTimeout:         30 * time.Second,
			},
			DispatcherConfig: &tcpserver.DispatcherConfig{
				QueueSize: 100,
				WorkerNum: 8,
			},
		},
		l,
	)
	if err != nil {
		return fmt.Errorf("Open tcp sever failed, %v", err)
	}
	glog.Info("Open tcp server OK.")

	return nil
}
