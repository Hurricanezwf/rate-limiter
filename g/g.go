package g

import (
	"flag"
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
	flag.StringVar(&ConfPath, "c", "./conf/config.yaml", "path of the config")

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
	if Config.Raft.Enable {
		if err := os.MkdirAll(Config.Raft.RootDir, os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}
