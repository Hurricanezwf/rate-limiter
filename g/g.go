package g

import (
	"flag"
	"fmt"
	"os"

	"github.com/Hurricanezwf/toolbox/logging"
)

var (
	// 配置路径
	ConfPath = "./conf/config.yaml"

	Config *Conf
)

func init() {
	flag.StringVar(&ConfPath, "c", "./conf/config.yaml", "path of the config")
}

func Init() (err error) {
	if Config, err = LoadConfig(ConfPath); err != nil {
		return fmt.Errorf("Load config failed, %v", err)
	}
	if err = initialize(); err != nil {
		return fmt.Errorf("Initialize failed, %v", err)
	}
	Config.Debug()
	return nil
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
