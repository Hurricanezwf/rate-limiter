package services

import (
	"fmt"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

// Rate limiter实例
var l limiter.Limiter = limiter.New()

func Run() (err error) {
	glog.Infof("Be starting, wait a while...")

	//
	if err = l.Open(); err != nil {
		return fmt.Errorf("Open limiter failed, %v", err)
	}
	glog.Info("Open limiter OK.")

	//
	if err := runHttpd(g.Config.Httpd.Listen); err != nil {
		return fmt.Errorf("Run HTTP Server on %s failed, %v", g.Config.Httpd.Listen, err)
	}
	glog.Infof("Run HTTP Service OK, listen at %s.", g.Config.Httpd.Listen)

	return nil
}
