package g

import (
	"os"

	"github.com/Hurricanezwf/toolbox/logging"
	"github.com/Hurricanezwf/toolbox/logging/glog"
)

var (
	// 配置路径
	ConfPath = "./conf/config.yaml"

	Config *Conf
)

func init() {
	var err error

	if Config, err = LoadConfig(ConfPath); err != nil {
		glog.Fatalf("Load config failed, %v", err)
	}

	if err := initialize(); err != nil {
		glog.Fatalf("Initialize failed, %v", err)
	}
}

func initialize() error {
	// 初始化日志
	if err := logging.Reset(Config.Log.Way, Config.Log.Dir, Config.Log.Verbose); err != nil {
		return err
	}

	// 准备必要目录
	if err := os.MkdirAll(Config.Raft.RootDir, 755); err != nil {
		return err
	}

	return nil
}
