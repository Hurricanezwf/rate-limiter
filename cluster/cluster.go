package cluster

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/meta"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/golang/protobuf/proto"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func init() {
	RegistBuilder("v2", newClusterV2)
}

// Interface 集群的抽象接口
type Interface interface {
	// Open 启用集群
	Open() error

	// Raft's FSM
	raftlib.FSM

	// IsLeader check if current node is raft leader
	IsLeader() bool

	// LeaderHTTPAddr return current leader's http listen
	LeaderHTTPAddr() string

	// RegistQuota 注册资源配额
	RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp

	// Borrow 申请一次执行资格，如果成功返回nil
	// expire 表示申请的资源的自动回收时间
	Borrow(r *APIBorrowReq) *APIBorrowResp

	// Return 归还执行资格，如果成功返回nil
	Return(r *APIReturnReq) *APIReturnResp

	// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
	// 返回回收的资源数量
	ReturnAll(r *APIReturnAllReq) *APIReturnAllResp
}

// Default 新建一个默认cluster实例
func Default() Interface {
	return New("v2")
}

// New 新建一个cluster接口实例
func New(name string) Interface {
	if builders == nil {
		return nil
	}
	if f := builders[name]; f != nil {
		return f()
	}
	return nil
}

// clusterV2 cluster.Interface的具体实现
type clusterV2 struct {
	// 成员锁
	gLock *sync.RWMutex

	// 元数据
	m meta.Interface

	// Raft实例
	raft *raftlib.Raft

	// Leader结点HTTP服务地址
	leaderHTTPAddr string

	// Raft超时时间
	raftTimeout time.Duration

	// 控制退出
	stopC chan struct{}
}

func newClusterV2() Interface {
	return &clusterV2{
		gLock:       &sync.RWMutex{},
		m:           meta.Default(),
		raftTimeout: time.Duration(g.Config.Raft.Timeout) * time.Millisecond,
		stopC:       make(chan struct{}),
	}
}

// Open 启动集群
func (c *clusterV2) Open() error {
	const (
		STMOpenRaft = iota
		STMStartCleaner
		STMStartLeaderWatcher
		STMExit
	)

	for st := STMOpenRaft; st != STMExit; {
		switch st {
		case STMOpenRaft:
			// Raft集群初始化
			if err := c.initRaftCluster(); err != nil {
				return err
			}
			st = STMStartLeaderWatcher

		case STMStartLeaderWatcher:
			// 定时输出谁是Leader & 服务注册
			if g.Config.Raft.Enable {
				go c.initRaftLeaderWatcher()
			}
			st = STMStartCleaner

		case STMStartCleaner:
			// 定时回收&重用资源
			go c.initMetaCleaner()
			st = STMExit
		}
	}

	return nil
}

// initRaftCluster 初始化启动raft集群
func (c *clusterV2) initRaftCluster() error {
	if g.Config.Raft.Enable == false {
		glog.V(1).Info("Raft is disabled")
		return nil
	}

	var (
		bind            = g.Config.Raft.Bind
		localID         = g.Config.Raft.LocalID
		clusterConfJson = g.Config.Raft.ClusterConfJson
		storage         = g.Config.Raft.Storage
		rootDir         = g.Config.Raft.RootDir
		tcpMaxPool      = g.Config.Raft.TCPMaxPool
		timeout         = time.Duration(g.Config.Raft.Timeout) * time.Millisecond
		bootstrap       = false
		dbPath          = ""
	)

	// 检测集群是否需要bootstrap
	switch storage {
	case g.RaftStorageMemory:
		bootstrap = true
	default:
		dbPath = filepath.Join(rootDir, "raft.db")
		if _, err := os.Lstat(dbPath); err == nil {
			bootstrap = false
		} else if os.IsNotExist(err) {
			bootstrap = true
		} else {
			return fmt.Errorf("Lstat failed, %v", err)
		}
	}

	// 加载集群配置
	clusterConf, err := raftlib.ReadConfigJSON(clusterConfJson)
	if err != nil {
		return fmt.Errorf("Load raft servers' config failed, %v", err)
	}

	raftConf := raftlib.DefaultConfig()
	raftConf.LocalID = raftlib.ServerID(localID)
	if g.Config.Log.Verbose < 5 {
		raftConf.LogOutput = nil
		raftConf.Logger = nil
	} else {
		raftConf.LogOutput = os.Stderr
	}

	// 创建Raft网络传输
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return err
	}
	transport, err := raftlib.NewTCPTransport(
		bind,               // bindAddr
		addr,               // advertise
		tcpMaxPool,         // maxPool
		timeout,            // timeout
		raftConf.LogOutput, // logOutput
	)
	if err != nil {
		return fmt.Errorf("Create tcp transport failed, %v", err)
	}

	// 创建Log Entry存储引擎
	// 此处的存储引擎的选择基本决定了接口的延迟。
	// 本地落地存储效率比较低，单个请求响应时间在50ms以上；如果换成内存存储的话，单个请求响应时间在2ms左右。
	logStore, stableStore, err := c.storageEngineFactory(storage, dbPath)
	if err != nil {
		return fmt.Errorf("Create storage instance failed, %v", err)
	}

	// 创建持久化快照存储引擎
	snapshotStore, err := raftlib.NewFileSnapshotStore(rootDir, 3, raftConf.LogOutput)
	if err != nil {
		return fmt.Errorf("Create file snapshot store failed, %v", err)
	}

	// 创建Raft实例
	raft, err := raftlib.NewRaft(
		raftConf,
		c,
		logStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return fmt.Errorf("Create raft instance failed, %v", err)
	}

	// 启动集群
	if bootstrap {
		future := raft.BootstrapCluster(clusterConf)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Bootstrap raft cluster failed, %v", err)
		}
	}

	// 变量初始化
	c.raft = raft

	return nil
}

// initRaftWatcher 初始化启动Raft Leader监听器
func (c *clusterV2) initRaftLeaderWatcher() {
	ticker := time.NewTicker(time.Duration(g.Config.Raft.LeaderWatchInterval) * time.Second)

	for {
		select {
		case <-c.stopC:
			return
		case <-ticker.C:
			glog.V(2).Infof("(L) %s  ==>  %s", c.raft.Leader(), c.LeaderHTTPAddr())
		case <-c.raft.LeaderCh():
			{
				glog.V(2).Infof("Leader changed, %s is Leader", c.raft.Leader())

				if c.IsLeader() == false {
					continue
				}

				// 通知所有结点到Leader上去服务注册
				cmd, err := encodeCMD(ActionLeaderNotify, &CMDLeaderNotify{
					RaftAddr: g.Config.Raft.Bind,
					HttpAddr: g.Config.Httpd.Listen,
				})
				if err != nil {
					glog.Warningf("Marshal ActionLeaderNotify failed, %v", err)
					continue
				}

				future := c.raft.Apply(cmd, c.raftTimeout)
				if err := future.Error(); err != nil {
					glog.Warningf("Raft apply CMDLeaderNotify failed, %v", err)
					continue
				}
				if future.Response() != nil {
					glog.Warningf("Raft Apply CMDBLeaderNotify failed, %v", future.Response())
					continue
				}
			}
		}
	}

}

// initMetaCleaner 周期性清理过期资源和重用已回收的资源
func (c *clusterV2) initMetaCleaner() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-c.stopC:
			return
		case <-ticker.C:
			c.m.Recycle()
		}
	}
}

// IsLeader 该结点在集群中是否是Leader
func (c *clusterV2) IsLeader() bool {
	if g.Config.Raft.Enable == false {
		return true
	}
	return c.raft.State() == raftlib.Leader
}

// LeaderHTTPAddr 返回Leader结点的HTTP服务地址
func (c *clusterV2) LeaderHTTPAddr() string {
	c.gLock.RLock()
	defer c.gLock.RUnlock()
	return c.leaderHTTPAddr
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (c *clusterV2) Apply(log *raftlib.Log) interface{} {
	switch log.Type {
	case raftlib.LogCommand:
		{
			cmd, args := resolveCMD(log.Data)
			start := time.Now()
			glog.V(3).Infof("Apply CMD: %#v", cmd)
			rt := c.switchDo(cmd, args)
			glog.V(3).Infof("Response : %+v, elapse:%v", rt, time.Since(start))
			return rt
		}
	default:
		return fmt.Errorf("Unknown LogType(%v)", log.Type)
	}
	return nil
}

// Snapshot 获取对FSM进行快照操作的实例
// Encode Format:
// > [1 byte]  magic number
// > [1 byte]  protocol version
// > [8 bytes] timestamp
// > [N bytes] data encoded bytes
func (c *clusterV2) Snapshot() (raftlib.FSMSnapshot, error) {
	start := time.Now()
	glog.V(1).Info("Create snapshot starting...")

	// 编码元数据
	b, err := c.metaBytes()
	if err != nil {
		glog.Warning(err.Error())
		return nil, err
	}

	// 构造快照格式
	buf := bytes.NewBuffer(make([]byte, 0, len(b)+10))

	if err = buf.WriteByte(MagicNumber); err != nil {
		glog.Warning(err.Error())
		return nil, err
	}
	if err = buf.WriteByte(ProtocolVersion); err != nil {
		glog.Warning(err.Error())
		return nil, err
	}
	if err = binary.Write(buf, binary.BigEndian, start.Unix()); err != nil {
		glog.Warning(err.Error())
		return nil, err
	}
	if _, err = buf.Write(b); err != nil {
		glog.Warning(err.Error())
		return nil, err
	}

	glog.V(1).Infof("Create snapshot finished, elapse:%v, totalSize:%d", time.Since(start), buf.Len())

	return NewSnapshot(buf), nil
}

// Restore 从快照中恢复LimiterFSM
func (c *clusterV2) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	glog.V(1).Info("Restore snapshot starting...")

	start := time.Now()
	reader := bufio.NewReader(rc)

	// 读取识别魔数
	magicNumber, err := reader.ReadByte()
	if err != nil {
		glog.Warningf("Read magic number failed, %v", err)
		return err
	}
	if magicNumber != MagicNumber {
		err = fmt.Errorf("Unknown magic number %#x", magicNumber)
		glog.Warning(err.Error())
		return err
	}

	// 读取识别协议版本
	protocolVersion, err := reader.ReadByte()
	if err != nil {
		glog.Warningf("Read protocol version failed, %v", err)
		return err
	}
	if protocolVersion != ProtocolVersion {
		err = fmt.Errorf("ProtocolVersion(%#x) is not match with %#x", protocolVersion, ProtocolVersion)
		glog.Warning(err.Error())
		return err
	}

	// 读取快照时间戳
	ts := make([]byte, 8)
	n, err := reader.Read(ts)
	if err != nil {
		glog.Warningf("Read timestamp failed, %v", err)
		return err
	}
	if n != len(ts) {
		err = errors.New("Read timestamp failed, missing data")
		glog.Warning(err.Error())
		return err
	}
	tsInt64 := binary.BigEndian.Uint64(ts)
	glog.V(2).Infof("Restore from snapshot created at %d", tsInt64)

	// 读取元数据并解析
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	bodyBytesCount, err := buf.ReadFrom(reader)
	if err != nil && err != io.EOF {
		glog.Warningf("Read meta data failed, %v", err)
		return err
	}

	m := meta.Default()
	if err = m.Decode(buf.Bytes()); err != nil {
		glog.Warningf("Decode meta data failed, %v", err)
		return err
	}

	// 替换元数据
	c.gLock.Lock()
	c.m = m
	c.gLock.Unlock()

	glog.V(1).Infof("Restore snapshot finished, elapse:%v, totalBytes:%d", time.Since(start), 10+bodyBytesCount)

	return nil
}

// RegistQuota 注册资源配额
func (c *clusterV2) RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp {
	rp := &APIRegistQuotaResp{Code: 500}
	cmd, err := encodeCMD(ActionRegistQuota, r)
	if err != nil {
		rp.Msg = fmt.Sprintf("Encode cmd failed, %v", err)
		return rp
	}

	if g.Config.Raft.Enable == false {
		_, args := resolveCMD(cmd)
		return c.handleRegistQuota(args)
	}

	future := c.raft.Apply(cmd, c.raftTimeout)
	if err = future.Error(); err != nil {
		rp.Msg = err.Error()
		return rp
	}
	return future.Response().(*APIRegistQuotaResp)
}

// Borrow 申请一次执行资格，如果成功返回nil
func (c *clusterV2) Borrow(r *APIBorrowReq) *APIBorrowResp {
	rp := &APIBorrowResp{Code: 500}
	cmd, err := encodeCMD(ActionBorrow, r)
	if err != nil {
		rp.Msg = fmt.Sprintf("Encode cmd failed, %v", err)
		return rp
	}

	if g.Config.Raft.Enable == false {
		_, args := resolveCMD(cmd)
		return c.handleBorrow(args)
	}

	future := c.raft.Apply(cmd, c.raftTimeout)
	if err = future.Error(); err != nil {
		rp.Msg = err.Error()
		return rp
	}
	return future.Response().(*APIBorrowResp)
}

// Return 归还执行资格，如果成功返回nil
func (c *clusterV2) Return(r *APIReturnReq) *APIReturnResp {
	rp := &APIReturnResp{Code: 500}
	cmd, err := encodeCMD(ActionReturn, r)
	if err != nil {
		rp.Msg = fmt.Sprintf("Encode cmd failed, %v", err)
		return rp
	}

	if g.Config.Raft.Enable == false {
		_, args := resolveCMD(cmd)
		return c.handleReturn(args)
	}

	future := c.raft.Apply(cmd, c.raftTimeout)
	if err = future.Error(); err != nil {
		rp.Msg = err.Error()
		return rp
	}
	return future.Response().(*APIReturnResp)
}

// ReturnAll 归还某个用户所有的执行资格，通常在用户主动关闭的时候
func (c *clusterV2) ReturnAll(r *APIReturnAllReq) *APIReturnAllResp {
	rp := &APIReturnAllResp{Code: 500}
	cmd, err := encodeCMD(ActionReturnAll, r)
	if err != nil {
		rp.Msg = fmt.Sprintf("Encode cmd failed, %v", err)
		return rp
	}

	if g.Config.Raft.Enable == false {
		_, args := resolveCMD(cmd)
		return c.handleReturnAll(args)
	}

	future := c.raft.Apply(cmd, c.raftTimeout)
	if err = future.Error(); err != nil {
		rp.Msg = err.Error()
		return rp
	}
	return future.Response().(*APIReturnAllResp)
}

// switchDo 根据命令类型选择执行
func (c *clusterV2) switchDo(cmd byte, args []byte) interface{} {
	switch cmd {
	case ActionRegistQuota:
		return c.handleRegistQuota(args)
	case ActionBorrow:
		return c.handleBorrow(args)
	case ActionReturn:
		return c.handleReturn(args)
	case ActionReturnAll:
		return c.handleReturnAll(args)
	case ActionLeaderNotify:
		return c.handleLeaderNotify(args)
	}
	glog.Fatalf("Unknown cmd '%#v'", cmd)
	return fmt.Errorf("Unknown cmd '%#v'", cmd)
}

// handleRegistQuota 处理注册资源配额的逻辑
func (c *clusterV2) handleRegistQuota(args []byte) *APIRegistQuotaResp {
	var err error
	var r APIRegistQuotaReq
	var rp APIRegistQuotaResp

	if err = resolveArgs(args, &r); err != nil {
		rp.Code = 500
		rp.Msg = fmt.Sprintf("Resolve request failed, %v", err)
		return &rp
	}

	if err = c.m.RegistQuota(r.RCType, r.Quota); err != nil {
		rp.Code = 403
		rp.Msg = err.Error()
		return &rp
	}

	return &rp
}

// handleBorrow 处理借取资源的逻辑
func (c *clusterV2) handleBorrow(args []byte) *APIBorrowResp {
	var err error
	var rcId string
	var r APIBorrowReq
	var rp APIBorrowResp

	if err = resolveArgs(args, &r); err != nil {
		rp.Code = 500
		rp.Msg = fmt.Sprintf("Resolve request failed, %v", err)
		return &rp
	}

	if rcId, err = c.m.Borrow(r.RCType, r.ClientID, r.Expire); err != nil {
		rp.Code = 403
		rp.Msg = err.Error()
		return &rp
	} else {
		rp.RCID = rcId
	}

	return &rp
}

// handleReturn 处理归还单个资源的逻辑
func (c *clusterV2) handleReturn(args []byte) *APIReturnResp {
	var err error
	var r APIReturnReq
	var rp APIReturnResp

	if err = resolveArgs(args, &r); err != nil {
		rp.Code = 500
		rp.Msg = fmt.Sprintf("Resolve request failed, %v", err)
		return &rp
	}

	if err = c.m.Return(r.ClientID, r.RCID); err != nil {
		rp.Code = 403
		rp.Msg = err.Error()
		return &rp
	}

	return &rp
}

// handleReturnAll 处理归还某用户占用的所有资源的逻辑
func (c *clusterV2) handleReturnAll(args []byte) *APIReturnAllResp {
	var err error
	var r APIReturnAllReq
	var rp APIReturnAllResp

	if err = resolveArgs(args, &r); err != nil {
		rp.Code = 500
		rp.Msg = fmt.Sprintf("Resolve request failed, %v", err)
		return &rp
	}
	if _, err = c.m.ReturnAll(r.RCType, r.ClientID); err != nil {
		rp.Code = 403
		rp.Msg = err.Error()
		return &rp
	}

	return &rp
}

// handleLeaderNotify 处理Leader结点通知所有结点服务地址的逻辑
func (c *clusterV2) handleLeaderNotify(args []byte) error {
	var err error
	var r CMDLeaderNotify

	if err = resolveArgs(args, &r); err != nil {
		return err
	}

	c.gLock.Lock()
	c.leaderHTTPAddr = r.HttpAddr
	c.gLock.Unlock()

	return err
}

// storageEngineFactory 根据存储引擎名字构造存储实例
func (c *clusterV2) storageEngineFactory(name string, dbPath ...string) (raftlib.LogStore, raftlib.StableStore, error) {
	var (
		err         error
		logStore    raftlib.LogStore
		stableStore raftlib.StableStore
	)

	switch name {
	case g.RaftStorageBoltDB:
		{
			if len(dbPath) < 1 {
				err = fmt.Errorf("Missing dbPath for storage '%s'", name)
				break
			}
			if storage, e := raftboltdb.NewBoltStore(dbPath[0]); e != nil {
				err = e
			} else {
				logStore = storage
				stableStore = storage
			}
		}
	case g.RaftStorageMemory:
		{
			storage := raftlib.NewInmemStore()
			logStore = storage
			stableStore = storage
		}
	default:
		err = fmt.Errorf("No storage engine found for %s", name)
	}

	return logStore, stableStore, err
}

// metaBytes 并发安全地将元数据序列化
func (c *clusterV2) metaBytes() ([]byte, error) {
	c.gLock.RLock()
	defer c.gLock.RUnlock()
	return c.m.Encode()
}

// encodeCMD 编码在Raft集群内执行的命令
func encodeCMD(cmd byte, args proto.Message) ([]byte, error) {
	b, err := proto.Marshal(args)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1+len(b)))
	buf.WriteByte(cmd)
	buf.Write(b)
	return buf.Bytes(), nil
}

// resolveCMD 解析出命令类型
func resolveCMD(b []byte) (cmd byte, args []byte) {
	if len(b) < 1 {
		return 0, nil
	}
	return b[0], b[1:]
}

// resolveArgs 解析出命令参数
func resolveArgs(args []byte, v proto.Message) error {
	return proto.Unmarshal(args, v)
}
