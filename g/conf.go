package g

import (
	"errors"
	"io/ioutil"
	"net"

	"github.com/Hurricanezwf/toolbox/logging"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	yaml "gopkg.in/yaml.v2"
)

const (
	RaftStorageMemory = "memory"
	RaftStorageBoltDB = "boltdb"
)

func DefaultConf() *Conf {
	return &Conf{
		//
		Log: &LogConf{
			Verbose: 2,
			Way:     logging.LogWayConsole,
		},

		//
		Httpd: &HttpdConf{
			Listen: "0.0.0.0:17250",
		},

		//
		Raft: &RaftConf{
			Enable:              true,
			Storage:             RaftStorageBoltDB,
			TCPMaxPool:          3,
			Timeout:             10000,
			LeaderWatchInterval: 60,
		},
	}
}

// 全局配置
type Conf struct {
	Log   *LogConf   `yaml:"Log"`
	Httpd *HttpdConf `yaml:"Httpd"`
	Raft  *RaftConf  `yaml:"Raft"`
}

func (c *Conf) Debug() {
	glog.V(1).Infof("-------------------- RATE LIMITER CONFIG ---------------------")
	glog.V(1).Infof("%-32s : %d", "Log.Verbose", c.Log.Verbose)
	glog.V(1).Infof("%-32s : %s", "Log.Way", c.Log.Way)
	glog.V(1).Infof("%-32s : %s", "Log.Dir", c.Log.Dir)
	glog.V(1).Infof("%-32s : %s", "Httpd.Listen", c.Httpd.Listen)
	glog.V(1).Infof("%-32s : %t", "Raft.Enable", c.Raft.Enable)
	glog.V(1).Infof("%-32s : %s", "Raft.LocalID", c.Raft.LocalID)
	glog.V(1).Infof("%-32s : %s", "Raft.Bind", c.Raft.Bind)
	glog.V(1).Infof("%-32s : %s", "Raft.ClusterConfJson", c.Raft.ClusterConfJson)
	glog.V(1).Infof("%-32s : %s", "Raft.Storage", c.Raft.Storage)
	glog.V(1).Infof("%-32s : %s", "Raft.RootDir", c.Raft.RootDir)
	glog.V(1).Infof("%-32s : %d", "Raft.TCP", c.Raft.TCPMaxPool)
	glog.V(1).Infof("%-32s : %d", "Raft.Timeout", c.Raft.Timeout)
	glog.V(1).Infof("%-32s : %d", "Raft.LeaderWatchInterval", c.Raft.LeaderWatchInterval)
	glog.V(1).Infof("--------------------------------------------------------------")
}

// Httpd配置
type HttpdConf struct {
	Listen string `yaml:"Listen"`
}

// Raft算法配置
type RaftConf struct {
	Enable              bool   `yaml:"Enable"`
	LocalID             string `yaml:"ID"`
	Bind                string `yaml:"Bind"`
	ClusterConfJson     string `yaml:"ClusterConfJson"`
	Storage             string `yaml:"Storage"`
	RootDir             string `yaml:"RootDir"`
	TCPMaxPool          int    `yaml:"TCPMaxPool"`
	Timeout             int64  `yaml:"Timeout"`             // Raft算法commit超时时间，单位毫秒
	LeaderWatchInterval int64  `yaml:"LeaderWatchInterval"` // 定时报告谁是Leader, 单位秒
}

// 日志配置
type LogConf struct {
	Verbose int    `yaml:"Verbose"` // 日志冗余度
	Way     string `yaml:"Way"`     // 日志方式"console" | "file"
	Dir     string `yaml:"Dir"`     // 日志目录，仅当Way是"file"是需要
}

func LoadConfig(path string) (*Conf, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	c := DefaultConf()
	if err = yaml.Unmarshal(b, c); err != nil {
		return nil, err
	}

	// 日志参数校验
	if c.Log.Verbose < 0 {
		c.Log.Verbose = 2
	}

	switch c.Log.Way {
	case logging.LogWayConsole:
		break
	case logging.LogWayFile:
		if len(c.Log.Dir) <= 0 {
			return nil, errors.New("Missing `Log.Dir` field")
		}
	default:
		return nil, errors.New("Invalid `Log.Way` field value")
	}

	// Httpd参数校验
	addr, err := net.ResolveTCPAddr("tcp", c.Httpd.Listen)
	if err != nil {
		return nil, errors.New("Invalid `Httpd.Listen` field value")
	}
	if addr.IP.IsUnspecified() {
		return nil, errors.New("Invalid `Httpd.Listen` field value, it is not advertisable")
	}

	// Raft参数校验
	if c.Raft.Enable {
		if _, err = net.ResolveTCPAddr("tcp", c.Raft.Bind); err != nil {
			return nil, errors.New("Invalid `Raft.Bind` field value")
		}

		if len(c.Raft.ClusterConfJson) <= 0 {
			return nil, errors.New("Missing `Raft.ClusterConfJson` field")
		}

		if len(c.Raft.LocalID) <= 0 {
			return nil, errors.New("Missing `Raft.LocalID` field")
		}

		if raftStorageOK(c.Raft.Storage) == false {
			return nil, errors.New("Invalid `Raft.Storage` field value")
		}

		if len(c.Raft.RootDir) <= 0 {
			return nil, errors.New("Missing `Raft.RootDir` field")
		}

		if c.Raft.TCPMaxPool <= 0 {
			return nil, errors.New("Missing `Raft.TCPMaxPool` field")
		}

		if c.Raft.Timeout <= 0 {
			return nil, errors.New("Missing `Raft.Timeout` field")
		}
	}

	return c, nil
}

func raftStorageOK(s string) bool {
	switch s {
	case RaftStorageBoltDB, RaftStorageMemory:
		return true
	}
	return false
}
