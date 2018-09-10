package version

import (
	"fmt"

	"github.com/Hurricanezwf/toolbox/logging/glog"
)

var (
	CompileTime string
	GoVersion   string
	Branch      string
	Commit      string
)

func Display() error {
	glog.V(1).Info()
	glog.V(1).Infof("CompileTime : %s", CompileTime)
	glog.V(1).Infof("GoVersion   : %s", GoVersion)
	glog.V(1).Infof("Branch      : %s", Branch)
	glog.V(1).Infof("Commit      : %s", Commit)
	glog.V(1).Info()
	return nil
}

func String() string {
	return fmt.Sprintf(`{"CompileTime":"%s","GoVersion":"%s","Branch":"%s","Commit":"%s"}`,
		CompileTime, GoVersion, Branch, Commit)
}
