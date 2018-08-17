package g

import "time"

var (
	// 配置目录
	ConfDir = "./conf"

	// RaftServer配置文件
	RaftClusterConfFile = "raft-cluster.json"

	// 存放Raft文件的目录，默认"./raft"
	RaftDir = "./raft"

	// Raft监听地址
	RaftBind string

	// Raft的TCP连接池最大数
	RaftTCPMaxPool = 3

	// Raft的超时时间
	RaftTimeout = 10 * time.Second
)
