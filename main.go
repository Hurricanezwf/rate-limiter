package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Hurricanezwf/rate-limiter/services"
	"github.com/Hurricanezwf/toolbox/logging"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

func main() {
	flag.Parse()
	defer glog.Flush()

	if err := logging.Reset(logging.LogWayConsole, "", 5); err != nil {
		glog.Fatalf(err.Error())
	}

	if err := services.Run(); err != nil {
		glog.Fatalf(err.Error())
	}

	// wait exit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	s := <-sig
	glog.Warningf("Receive %s, so exit", s)
}
