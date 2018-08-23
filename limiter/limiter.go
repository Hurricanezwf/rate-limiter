package limiter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	"github.com/Hurricanezwf/rate-limiter/g"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// New new one default limiter
func New() Limiter {
	return &limiterV1{
		mutex:        &sync.RWMutex{},
		meta:         encoding.NewMap(),
		serviceMutex: &sync.RWMutex{},
		services:     make(map[string]*APIRegistServiceReq),
		stopC:        make(chan struct{}),
	}
}

// Limiter is an abstract of rate limiter
type Limiter interface {
	//
	Open() error

	//
	Do(r *Request) *Response

	//
	raftlib.FSM

	//
	IsLeader() bool
}

// limiterV1 is an implement of Limiter interface
// It use raft to ensure consistency among all nodes
type limiterV1 struct {
	// 读写锁保护meta
	mutex *sync.RWMutex

	// 按照资源类型进行分类
	// ResourceTypeHex ==> LimiterMeta
	meta *encoding.Map

	// Raft实例
	raft *raftlib.Raft

	// Raft地址到Httpd服务地址的映射关系
	// raftBind ==> *APIRegistServiceReq
	serviceMutex *sync.RWMutex
	services     map[string]*APIRegistServiceReq

	// 控制退出
	stopC chan struct{}
}

func (l *limiterV1) Open() error {
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
func (l *limiterV1) initRaftCluster() error {
	if g.Config.Raft.Enable == false {
		glog.Info("Raft is disabled")
		return nil
	}

	var (
		bind            = g.Config.Raft.Bind
		localID         = g.Config.Raft.LocalID
		clusterConfJson = g.Config.Raft.ClusterConfJson
		rootDir         = g.Config.Raft.RootDir
		tcpMaxPool      = g.Config.Raft.TCPMaxPool
		timeout         = time.Duration(g.Config.Raft.Timeout) * time.Millisecond
	)

	// 加载集群配置
	clusterConf, err := raftlib.ReadConfigJSON(clusterConfJson)
	if err != nil {
		return fmt.Errorf("Load raft servers' config failed, %v", err)
	}

	raftConf := raftlib.DefaultConfig()
	raftConf.LocalID = raftlib.ServerID(localID)

	// 创建Raft网络传输
	addr, err := net.ResolveTCPAddr("tcp", bind)
	if err != nil {
		return err
	}
	transport, err := raftlib.NewTCPTransport(
		bind,       // bindAddr
		addr,       // advertise
		tcpMaxPool, // maxPool
		timeout,    // timeout
		os.Stderr,  // logOutput
	)
	if err != nil {
		return fmt.Errorf("Create tcp transport failed, %v", err)
	}

	// 创建持久化Log Entry存储引擎
	dbPath := filepath.Join(rootDir, "raft.db")
	boltDB, err := raftboltdb.NewBoltStore(dbPath)
	if err != nil {
		return fmt.Errorf("Create log store failed, %v", err)
	}

	// 创建持久化快照存储引擎
	snapshotStore, err := raftlib.NewFileSnapshotStore(rootDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf("Create file snapshot store failed, %v", err)
	}

	// 创建Raft实例
	raft, err := raftlib.NewRaft(
		raftConf,
		l,
		boltDB,
		boltDB,
		snapshotStore,
		transport,
	)
	if err != nil {
		return fmt.Errorf("Create raft instance failed, %v", err)
	}

	// 启动集群
	if PathExists(dbPath) == false {
		future := raft.BootstrapCluster(clusterConf)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Bootstrap raft cluster failed, %v", err)
		}
	} else {
		// no need to bootstrap
	}

	// 变量初始化
	l.raft = raft

	return nil
}

// initRaftWatcher 初始化启动Raft Leader监听器
func (l *limiterV1) initRaftLeaderWatcher() {
	ticker := time.NewTicker(time.Duration(g.Config.Raft.LeaderWatchInterval) * time.Second)

	for {
		select {
		case <-l.stopC:
			return
		case <-ticker.C:
			glog.V(2).Info("Registed Service List:")
			for raftAddr, httpdAddr := range l.Services() {
				if l.raft.Leader() == raftlib.ServerAddress(raftAddr) {
					glog.V(2).Infof("(L) %s  ==>  %s", raftAddr, httpdAddr)
				} else {
					glog.V(2).Infof("(F) %s  ==>  %s", raftAddr, httpdAddr)
				}
			}
		case <-l.raft.LeaderCh():
			{
				glog.V(2).Infof("Leader changed, %s is Leader", l.raft.Leader())

				if l.IsLeader() == false {
					continue
				}

				// 通知所有结点到Leader上去服务注册
				cmd := &Request{
					Action: ActionRegistServiceBroadcast,
					RegistServiceBroadcast: &CMDRegistServiceBroadcast{
						Timestamp: time.Now().UnixNano(),
						LeaderUrl: fmt.Sprintf("http://%s/v1/service/regist", g.Config.Httpd.Listen),
					},
				}

				b, err := json.Marshal(cmd)
				if err != nil {
					glog.Warningf("Marshal ActionRegistServiceBroadcast failed, %v", err)
					continue
				}
				glog.Info(string(b))

				future := l.raft.Apply(b, time.Duration(g.Config.Raft.Timeout)*time.Millisecond)
				if err := future.Error(); err != nil {
					glog.Warningf("Raft apply ActionRegistServiceBroadcast failed, %v", err)
					continue
				}
			}
		}
	}

}

func (l *limiterV1) initMetaCleaner() {
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

func (l *limiterV1) IsLeader() bool {
	if g.Config.Raft.Enable == false {
		return true
	}
	return l.raft.State() == raftlib.Leader
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (l *limiterV1) Apply(log *raftlib.Log) interface{} {
	switch log.Type {
	case raftlib.LogCommand:
		{
			// 解析远端传过来的命令日志
			glog.Info(string(log.Data))
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
// > [9 bytes] timestamp
// > [N bytes] data encoded bytes
func (l *limiterV1) Snapshot() (raftlib.FSMSnapshot, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	start := time.Now()
	glog.V(1).Info("Create snapshot starting...")

	// 编码元数据
	b, err := l.meta.Encode()
	if err != nil {
		glog.Warning(err.Error())
		return nil, err
	}

	// 编码时间戳
	ts, err := encoding.NewInt64(start.Unix()).Encode()
	if err != nil {
		glog.Warning(err.Error())
		return nil, err
	}

	// 构造快照
	buf := bytes.NewBuffer(make([]byte, 0, 10+len(b)))
	buf.WriteByte(MagicNumber)
	buf.WriteByte(ProtocolVersion)
	buf.Write(ts)
	buf.Write(b)

	//_ = ts
	//_ = b

	//f, err := os.OpenFile("./snapshot.limiter", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	//if err != nil {
	//	panic(err.Error())
	//}
	//defer f.Close()

	//if _, err = f.Write(buf.Bytes()); err != nil {
	//	panic(err.Error())
	//}

	//glog.V(2).Infof("%#v", buf.Bytes())

	glog.V(1).Infof("Create snapshot finished, elapse:%v, totalSize:%d", time.Since(start), buf.Len())

	return NewLimiterSnapshot(buf), nil
}

// Restore 从快照中恢复LimiterFSM
func (l *limiterV1) Restore(rc io.ReadCloser) error {
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
	ts := make([]byte, 9)
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
	tsInt64 := encoding.NewInt64(0)
	if _, err = tsInt64.Decode(ts); err != nil {
		err = fmt.Errorf("Decode timestamp failed, %v", err)
		glog.Warning(err.Error())
		return err
	}
	glog.V(1).Infof("Restore from snapshot created at %d", tsInt64.Value())

	// 读取元数据并解析
	buf := bytes.NewBuffer(make([]byte, 0, 10240))
	bodyBytesCount, err := buf.ReadFrom(reader)
	if err != nil && err != io.EOF {
		glog.Warningf("Read meta data failed, %v", err)
		return err
	}

	meta := encoding.NewMap()
	if _, err = meta.Decode(buf.Bytes()); err != nil {
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
func (l *limiterV1) Do(r *Request) (rp *Response) {
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

func (l *limiterV1) switchDo(r *Request) (rp *Response) {
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
	case ActionRegistServiceBroadcast:
		rp.Err = l.doRegistServiceBroadcast(r.RegistServiceBroadcast)
	case ActionRegistService:
		rp.Err = l.doRegistService(r.RegistService)
	default:
		rp.Err = fmt.Errorf("Action handler not found for %#v", r.Action)
	}

	if rp.Err != nil {
		glog.Warningf("Limiter: %v", rp.Err)
	}

	return rp
}

// doRegistQuota 注册资源配额
func (l *limiterV1) doRegistQuota(r *APIRegistQuotaReq) error {
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
	if _, existed := l.meta.Get(rcTypeIdHex); existed {
		return ErrExisted
	} else {
		l.meta.Set(rcTypeIdHex, NewLimiterMeta(r.RCTypeID, r.Quota))
	}

	glog.V(1).Infof("Regist quota for rcType '%s' OK, total:%d", rcTypeIdHex, r.Quota)

	return nil
}

// doBorrow 借一个资源返回
func (l *limiterV1) doBorrow(r *APIBorrowReq) (*APIBorrowResp, error) {
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
	m, ok := l.meta.Get(rcTypeIdHex)
	if !ok {
		return nil, fmt.Errorf("ResourceType(%s) is not registed", rcTypeIdHex)
	}
	rcId, err := m.(LimiterMeta).Borrow(r.ClientID, r.Expire)
	if err != nil {
		return nil, err
	}

	return &APIBorrowResp{RCID: rcId}, nil
}

func (l *limiterV1) doReturn(r *APIReturnReq) error {
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

	rcTypeId, err := ResolveResourceID(r.RCID)
	if err != nil {
		return fmt.Errorf("Resolve resource id failed, %v", err)
	}

	rcTypeIdHex := encoding.BytesToStringHex(rcTypeId)
	m, ok := l.meta.Get(rcTypeIdHex)
	if !ok {
		return fmt.Errorf("ResourceType(%s) is not registed", rcTypeIdHex)
	}
	if err := m.(LimiterMeta).Return(r.ClientID, r.RCID); err != nil {
		return err
	}

	return nil
}

func (l *limiterV1) doReturnAll(r *APIReturnAllReq) error {
	// check required
	if len(r.ClientID) <= 0 {
		return errors.New("Missing `clientId` field value")
	}
	clientIdHex := encoding.BytesToStringHex(r.ClientID)

	// try return all
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	quitC := make(chan struct{})
	defer close(quitC)

	for pair := range l.meta.Range(quitC) {
		rcTypeId := pair.K
		m := pair.V.(LimiterMeta)
		n, err := m.ReturnAll(r.ClientID)
		if err != nil {
			return fmt.Errorf("Client[%s] return all resource of type '%s' failed, %v", clientIdHex, rcTypeId, err)
		}
		glog.V(3).Infof("Client[%s] return all of type '%s' success, count:%d", clientIdHex, rcTypeId, n)
	}

	return nil
}

// 请求Leader地址注册服务
func (l *limiterV1) doRegistServiceBroadcast(r *CMDRegistServiceBroadcast) error {
	// check required
	if r.Timestamp <= 0 {
		return errors.New("Missing `timestamp` field value")
	}
	if len(r.LeaderUrl) <= 0 {
		return errors.New("Missing `leaderUrl` field value")
	}

	// 向Leader发起服务注册请求
	req := &Request{
		Action: ActionRegistService,
		RegistService: &APIRegistServiceReq{
			Timestamp: r.Timestamp,
			RaftAddr:  g.Config.Raft.Bind,
			HttpdAddr: g.Config.Httpd.Listen,
		},
	}

	b, err := json.Marshal(req)
	if err != nil {
		return err
	}

	rp, err := http.Post(r.LeaderUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		glog.Warning(err.Error())
		return err
	}
	defer rp.Body.Close()

	if rp.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(rp.Body)
		err := fmt.Errorf("Status(%d) != 200, %s", rp.StatusCode, string(msg))
		glog.Warning(err.Error())
		return err
	}

	return nil
}

// 注册服务
func (l *limiterV1) doRegistService(r *APIRegistServiceReq) error {
	// check required
	if r.Timestamp <= 0 {
		return errors.New("Missing `timestamp` field value")
	}
	if len(r.RaftAddr) <= 0 {
		return errors.New("Missing `raftAddr` field value")
	}
	if len(r.HttpdAddr) <= 0 {
		return errors.New("Missing `httpdAddr` field value")
	}

	l.serviceMutex.Lock()
	defer l.serviceMutex.Unlock()

	for _, d := range l.services {
		if d.Timestamp > r.Timestamp {
			// 滞后请求直接拒绝
			return errors.New("RegistServiceReq is old")
		} else if d.Timestamp < r.Timestamp {
			// 发现了之前注册的服务则全部丢弃
			l.services = make(map[string]*APIRegistServiceReq)
		} else {
			// 所有已注册的服务时间一致，按兵不动
			continue
		}
	}

	l.services[r.RaftAddr] = r
	glog.Info("[%d] Regist Service  %s  ==>  %s   OK", r.Timestamp, r.RaftAddr, r.HttpdAddr)

	return nil
}

func (l *limiterV1) recycle() {
	// map内容没有被修改，所以这里是读锁
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	quitC := make(chan struct{})
	defer close(quitC)

	for pair := range l.meta.Range(quitC) {
		limiterMeta := pair.V.(LimiterMeta)
		limiterMeta.Recycle()
	}
}

// Services 列出当前注册的所有服务
func (l *limiterV1) Services() map[string]string {
	l.serviceMutex.RLock()
	defer l.serviceMutex.RUnlock()

	ret := make(map[string]string)
	for _, d := range l.services {
		ret[d.RaftAddr] = d.HttpdAddr
	}
	return ret
}
