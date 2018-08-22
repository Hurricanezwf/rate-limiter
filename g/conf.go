package g

import (
	"errors"
	"io/ioutil"
	"net"

	"github.com/Hurricanezwf/toolbox/logging"
	yaml "gopkg.in/yaml.v2"
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
			Enable:     true,
			TCPMaxPool: 3,
			Timeout:    10000,
		},
	}
}

// 全局配置
type Conf struct {
	Log   *LogConf   `yaml:"Log"`
	Httpd *HttpdConf `yaml:"Httpd"`
	Raft  *RaftConf  `yaml:"Raft"`
}

// Httpd配置
type HttpdConf struct {
	Listen string `yaml:"Listen"`
}

// Raft算法配置
type RaftConf struct {
	Enable          bool   `yaml:"Enable"`
	LocalID         string `yaml:"ID"`
	Bind            string `yaml:"Bind"`
	ClusterConfJson string `yaml:"ClusterConfJson"`
	RootDir         string `yaml:"RootDir"`
	TCPMaxPool      int    `yaml:"TCPMaxPool"`
	Timeout         int64  `yaml:"Timeout"` // Raft算法commit超时时间，单位毫秒
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
	if _, err = net.ResolveTCPAddr("tcp", c.Httpd.Listen); err != nil {
		return nil, errors.New("Invalid `Httpd.Listen` field value")
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
