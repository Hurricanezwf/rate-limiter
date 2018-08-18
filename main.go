package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hurricanezwf/rate-limiter/server"
	"github.com/Hurricanezwf/toolbox/logging"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

var addr = "localhost:17250"

func main() {
	flag.Parse()
	defer glog.Flush()

	glog.Infof("Be starting, wait a while...")

	if err := logging.Reset(logging.LogWayConsole, "", 5); err != nil {
		glog.Fatalf(err.Error())
	}

	if err := server.Run(addr); err != nil {
		glog.Fatalf(err.Error())
	}

	glog.Infof("Start OK, listen at %s.", addr)

	// wait exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	glog.Warningf("Receive %s, so exit", s)
}
