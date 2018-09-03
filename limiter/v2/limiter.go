package limiterv1

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/meta"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/rate-limiter/types"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

func init() {
	limiter.RegistLimiterBuilder("v2", newLimiterV2)
}

// limiterV2 is an implement of Limiter interface
// It use raft to ensure consistency among all nodes
type limiterV2 struct {
	// 读写锁保护meta
	mutex *sync.RWMutex

	// 按照资源类型进行分类
	// ResourceTypeHex ==> LimiterMeta
	meta *types.PB_Map

	// Raft实例
	raft *raftlib.Raft

	// Leader结点的HTTP服务地址
	leaderHTTPAddr string

	// 控制退出
	stopC chan struct{}
}

func newLimiterV2() limiter.Limiter {
	return &limiterV2{
		mutex: &sync.RWMutex{},
		meta:  types.NewMap(),
		stopC: make(chan struct{}),
	}
}

func (l *limiterV2) Open() error {
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
			if err := l.initRaftCluster(); err != nil {
				return err
			}
			st = STMStartLeaderWatcher

		case STMStartLeaderWatcher:
			// 定时输出谁是Leader & 服务注册
			if g.Config.Raft.Enable {
				go l.initRaftLeaderWatcher()
			}
			st = STMStartCleaner

		case STMStartCleaner:
			// 定时回收&重用资源
			go l.initMetaCleaner()
			st = STMExit
		}
	}

	return nil
}

// initRaftCluster 初始化启动raft集群
func (l *limiterV2) initRaftCluster() error {
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
	)

	// 检测集群是否需要bootstrap
	dbPath := filepath.Join(rootDir, "raft.db")
	if _, err := os.Lstat(dbPath); err == nil {
		bootstrap = false
	} else if os.IsNotExist(err) {
		bootstrap = true
	} else {
		return fmt.Errorf("Lstat failed, %v", err)
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
	logStore, stableStore, err := l.storageEngineFactory(storage)
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
		l,
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
	l.raft = raft

	return nil
}

// initRaftWatcher 初始化启动Raft Leader监听器
func (l *limiterV2) initRaftLeaderWatcher() {
	ticker := time.NewTicker(time.Duration(g.Config.Raft.LeaderWatchInterval) * time.Second)

	for {
		select {
		case <-l.stopC:
			return
		case <-ticker.C:
			glog.V(2).Infof("(L) %s  ==>  %s", l.raft.Leader(), l.LeaderHTTPAddr())
		case <-l.raft.LeaderCh():
			{
				glog.V(2).Infof("Leader changed, %s is Leader", l.raft.Leader())

				if l.IsLeader() == false {
					continue
				}

				// 通知所有结点到Leader上去服务注册
				cmd := &Request{
					Action: ActionLeaderNotify,
					LeaderNotify: &CMDLeaderNotify{
						RaftAddr:  g.Config.Raft.Bind,
						HttpdAddr: g.Config.Httpd.Listen,
					},
				}

				b, err := json.Marshal(cmd)
				if err != nil {
					glog.Warningf("Marshal ActionLeaderNotify failed, %v", err)
					continue
				}

				future := l.raft.Apply(b, time.Duration(g.Config.Raft.Timeout)*time.Millisecond)
				if err := future.Error(); err != nil {
					glog.Warningf("Raft apply CMDLeaderNotify failed, %v", err)
					continue
				}
			}
		}
	}

}

func (l *limiterV2) initMetaCleaner() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-l.stopC:
			return
		case <-ticker.C:
			l.recycle()
		}
	}
}

func (l *limiterV2) IsLeader() bool {
	if g.Config.Raft.Enable == false {
		return true
	}
	return l.raft.State() == raftlib.Leader
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (l *limiterV2) Apply(log *raftlib.Log) interface{} {
	glog.V(4).Infof("Apply Log: %s", string(log.Data))

	switch log.Type {
	case raftlib.LogCommand:
		{
			// 解析远端传过来的命令日志
			var r Request
			if err := json.Unmarshal(log.Data, &r); err != nil {
				return fmt.Errorf("Bad request format for raft command, %v", err)
			}

			// 尝试执行命令
			return l.switchDo(&r)
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
func (l *limiterV2) Snapshot() (raftlib.FSMSnapshot, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	start := time.Now()
	glog.V(1).Info("Create snapshot starting...")

	// 编码元数据
	b, err := proto.Marshal(l.meta)
	if err != nil {
		glog.Warning(err.Error())
		return nil, err
	}

	// 编码时间戳
	ts := make([]byte, 8)
	binary.BigEndian.PutUint64(ts, uint64(start.Unix()))

	// 构造快照
	buf := bytes.NewBuffer(make([]byte, 0, 10+len(b)))
	buf.WriteByte(MagicNumber)
	buf.WriteByte(ProtocolVersion)
	buf.Write(ts)
	buf.Write(b)

	glog.V(1).Infof("Create snapshot finished, elapse:%v, totalSize:%d", time.Since(start), buf.Len())

	return limiter.NewLimiterSnapshot(buf), nil
}

// Restore 从快照中恢复LimiterFSM
func (l *limiterV2) Restore(rc io.ReadCloser) error {
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
	glog.V(1).Infof("Restore from snapshot created at %d", tsInt64)

	// 读取元数据并解析
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	bodyBytesCount, err := buf.ReadFrom(reader)
	if err != nil && err != io.EOF {
		glog.Warningf("Read meta data failed, %v", err)
		return err
	}

	meta := types.NewMap()
	if err := proto.Unmarshal(buf.Bytes(), meta); err != nil {
		glog.Warningf("Decode meta data failed, %v", err)
		return err
	}

	// 替换元数据
	l.mutex.Lock()
	l.meta = meta
	l.mutex.Unlock()

	glog.V(1).Infof("Restore snapshot finished, elapse:%v, totalBytes:%d", time.Since(start), 10+bodyBytesCount)

	return nil
}

// Do 执行请求
func (l *limiterV2) Do(r *Request) (rp *Response) {
	if g.Config.Raft.Enable == false {
		return l.switchDo(r)
	}

	cmd, err := json.Marshal(r)
	if err != nil {
		return &Response{Err: err}
	}

	future := l.raft.Apply(cmd, time.Duration(g.Config.Raft.Timeout)*time.Millisecond)
	if err = future.Error(); err != nil {
		return &Response{Err: fmt.Errorf("Raft apply error, %v", err)}
	}

	return future.Response().(*Response)
}

func (l *limiterV2) switchDo(r *Request) (rp *Response) {
	rp = &Response{Action: r.Action}
	switch r.Action {
	case ActionBorrow:
		rp.Borrow, rp.Err = l.doBorrow(r.Borrow)
	case ActionReturn:
		rp.Err = l.doReturn(r.Return)
	case ActionReturnAll:
		rp.Err = l.doReturnAll(r.ReturnAll)
	case ActionRegistQuota:
		rp.Err = l.doRegistQuota(r.RegistQuota)
	case ActionLeaderNotify:
		rp.Err = l.doLeaderNotify(r.LeaderNotify)
	default:
		rp.Err = fmt.Errorf("Action handler not found for %#v", r.Action)
	}

	if rp.Err != nil {
		glog.Warningf("Limiter: %v", rp.Err)
	}

	return rp
}

// doRegistQuota 注册资源配额
func (l *limiterV2) doRegistQuota(r *APIRegistQuotaReq) error {
	// check required
	if len(r.RCTypeID) <= 0 {
		return errors.New("Missing `rcTypeId` field value")
	}
	if r.Quota <= 0 {
		return errors.New("Invalid `quota` value")
	}

	// try regist
	l.mutex.Lock()
	defer l.mutex.Unlock()

	rcTypeIdHex := encoding.BytesToStringHex(r.RCTypeID)
	if v := l.meta.Get(rcTypeIdHex); v != nil {
		return ErrExisted
	} else {
		l.meta.Set(rcTypeIdHex, meta.New("v2", r.RCTypeID, r.Quota))
	}

	glog.V(1).Infof("Regist quota for rcType '%s' OK, total:%d", rcTypeIdHex, r.Quota)

	return nil
}

// doBorrow 借一个资源返回
func (l *limiterV2) doBorrow(r *APIBorrowReq) (*APIBorrowResp, error) {
	// check required
	if len(r.RCTypeID) <= 0 {
		return nil, errors.New("Missing `rcTypeId` field value")
	}
	if len(r.ClientID) <= 0 {
		return nil, errors.New("Missing `cId` field value")
	}
	if r.Expire <= 0 {
		return nil, errors.New("Missing `expire` field value")
	}

	// try borrow
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	rcTypeIdHex := encoding.BytesToStringHex(r.RCTypeID)
	m := l.meta.Get(rcTypeIdHex)
	if m == nil {
		return nil, fmt.Errorf("ResourceType[%s] is not registed", rcTypeIdHex)
	}

	limiterMeta := meta.New("v2", nil, 0)
	if err := types.UnmarshalAny(m, &limiterMeta); err != nil {
		return nil, fmt.Errorf("Unmarshal limiterMeta failed, %v", err)
	}
	rcId, err := m.(limiter.LimiterMeta).Borrow(r.ClientID, r.Expire)
	if err != nil {
		return nil, err
	}

	return &APIBorrowResp{RCID: rcId}, nil
}

func (l *limiterV2) doReturn(r *APIReturnReq) error {
	// check required
	if len(r.RCID) <= 0 {
		return errors.New("Missing `rcId` field value")
	}
	if len(r.ClientID) <= 0 {
		return errors.New("Missing `clientId` field value")
	}

	// try return
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	rcTypeId, err := limiter.ResolveResourceID(r.RCID)
	if err != nil {
		return fmt.Errorf("Resolve resource id failed, %v", err)
	}

	rcTypeIdHex := encoding.BytesToStringHex(rcTypeId)
	m, ok := l.meta.Get(rcTypeIdHex)
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", rcTypeIdHex)
	}
	if err := m.(limiter.LimiterMeta).Return(r.ClientID, r.RCID); err != nil {
		return err
	}

	return nil
}

func (l *limiterV2) doReturnAll(r *APIReturnAllReq) error {
	// check required
	if len(r.ClientID) <= 0 {
		return errors.New("Missing `clientId` field value")
	}
	clientIdHex := encoding.BytesToStringHex(r.ClientID)

	// try return all
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for rcTypeId, meta := range l.meta.Value {
		m := limiter.NewLimiterMeta("v2", nil, 0)
		if err := ptypes.UnmarshalAny(meta, &m); err != nil {
			glog.Warning(err.Error())
			return err
		}

		n, err := m.ReturnAll(r.ClientID)
		if err != nil {
			return fmt.Errorf("Client[%s] return all resource of type '%s' failed, %v", clientIdHex, rcTypeId, err)
		}
		glog.V(3).Infof("Client[%s] return all of type '%s' success, count:%d", clientIdHex, rcTypeId, n)
	}

	return nil
}

// doLeaderNotify save leader service info
func (l *limiterV2) doLeaderNotify(r *CMDLeaderNotify) error {
	// check required
	if len(r.RaftAddr) <= 0 {
		return errors.New("Missing `raftAddr` field value")
	}
	if len(r.HttpdAddr) <= 0 {
		return errors.New("Missing `httpdAddr` field value")
	}
	if raftlib.ServerAddress(r.RaftAddr) != l.raft.Leader() {
		return errors.New("Leader not consistent")
	}

	l.mutex.Lock()
	l.leaderHTTPAddr = r.HttpdAddr
	l.mutex.Unlock()

	return nil
}

func (l *limiterV2) recycle() {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	for rcTypeId, meta := range l.meta.Value {
		m := limiter.NewLimiterMeta("v2", nil, 0)
		if err := ptypes.UnmarshalAny(meta, &m); err != nil {
			glog.Warning(err.Error())
			continue
		}
		m.Recycle()
	}
}

func (l *limiterV2) LeaderHTTPAddr() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.leaderHTTPAddr
}

// storageEngineFactory 根据存储引擎名字构造存储实例
func (l *limiterV2) storageEngineFactory(name string, dbPath ...string) (raftlib.LogStore, raftlib.StableStore, error) {
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
