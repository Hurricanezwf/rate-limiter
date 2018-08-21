package limiter

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
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	raftlib "github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

// New new one default limiter
func New() Limiter {
	return &limiterV1{
		mutex: &sync.RWMutex{},
		meta:  encoding.NewMap(),
		stopC: make(chan struct{}),
	}
}

// Limiter is an abstract of rate limiter
type Limiter interface {
	//
	Open() error

	//
	Do(r *Request, commit bool) *Response

	//
	raftlib.FSM
}

// limiterV1 is an implement of Limiter interface
// It use raft to ensure consistency among all nodes
type limiterV1 struct {
	// 读写锁保护meta
	mutex *sync.RWMutex

	// 按照资源类型进行分类
	// ResourceTypeHex ==> LimiterMeta
	meta *encoding.Map

	//
	raft *raftlib.Raft

	//
	stopC chan struct{}
}

func (l *limiterV1) Open() error {
	// ---------------- Raft集群初始化 ----------------- //
	if g.EnableRaft {
		// 加载集群配置
		confJson := filepath.Join(g.ConfDir, g.RaftClusterConfFile)
		clusterConf, err := raftlib.ReadConfigJSON(confJson)
		if err != nil {
			return fmt.Errorf("Load raft servers' config failed, %v", err)
		}

		raftConf := raftlib.DefaultConfig()
		raftConf.LocalID = raftlib.ServerID(g.LocalID)

		// 创建Raft网络传输
		addr, err := net.ResolveTCPAddr("tcp", g.RaftBind)
		if err != nil {
			return err
		}
		transport, err := raftlib.NewTCPTransport(
			g.RaftBind,       // bindAddr
			addr,             // advertise
			g.RaftTCPMaxPool, // maxPool
			g.RaftTimeout,    // timeout
			os.Stderr,        // logOutput
		)
		if err != nil {
			return fmt.Errorf("Create tcp transport failed, %v", err)
		}

		// 创建持久化Log Entry存储引擎
		boltDB, err := raftboltdb.NewBoltStore(filepath.Join(g.RaftDir, "raft.db"))
		if err != nil {
			return fmt.Errorf("Create log store failed, %v", err)
		}

		// 创建持久化快照存储引擎
		snapshotStore, err := raftlib.NewFileSnapshotStore(g.RaftDir, 3, os.Stderr)
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
		future := raft.BootstrapCluster(clusterConf)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Bootstrap raft cluster failed, %v", err)
		}

		// 变量初始化
		l.raft = raft
	} else {
		glog.Info("Raft is disabled")
	}

	// ---------------------- 定时回收&重用资源 ---------------------- //
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-l.stopC:
				return
			case <-ticker.C:
				l.recycle()
			}
		}
	}()

	return nil
}

func (l *limiterV1) IsLeader() bool {
	if g.EnableRaft == false {
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
			var r Request
			if err := json.Unmarshal(log.Data, &r); err != nil {
				return fmt.Errorf("Bad request format for raft command, %v", err)
			}

			// 尝试执行命令
			if rp := l.Do(&r, false); rp.Err != nil {
				return fmt.Errorf("Try apply command failed, %v", rp.Err)
			}
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
	//buf.WriteByte(MagicNumber)
	//buf.WriteByte(ProtocolVersion)
	//buf.Write(ts)
	buf.Write(b)

	_ = ts

	f, err := os.OpenFile("./snapshot.limiter", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	if _, err = f.Write(buf.Bytes()); err != nil {
		panic(err.Error())
	}

	glog.V(2).Infof("%#v", buf.Bytes())

	glog.V(1).Infof("Create snapshot finished, elapse:%v, totalSize:%d", time.Since(start), buf.Len())

	return NewLimiterSnapshot(buf.Bytes()), nil
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
		err = fmt.Errorf("Unknown magic number %x", magicNumber)
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
		err = fmt.Errorf("ProtocolVersion(%x) is not match with %x", protocolVersion, ProtocolVersion)
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
	glog.V(1).Infof("Restore from snapshot created at %d", binary.BigEndian.Uint64(ts))

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
func (l *limiterV1) Do(r *Request, commit bool) (rp *Response) {
	rp = &Response{}
	switch r.Action {
	case ActionRegistQuota:
		rp.Err = l.doRegistQuota(r.RegistQuota, commit)
	case ActionBorrow:
		rp.Borrow, rp.Err = l.doBorrow(r.Borrow, commit)
	case ActionReturn:
		rp.Err = l.doReturn(r.Return, commit)
	case ActionDead:
		rp.Err = l.doReturnAll(r.Disconnect, commit)
	default:
		rp.Err = errors.New("Action handler not found")
	}

	if rp.Err != nil {
		glog.Warningf("Limiter: %v", rp.Err)
	}

	rp.Action = r.Action
	return rp
}

// doRegistQuota 注册资源配额
func (l *limiterV1) doRegistQuota(r *APIRegistQuotaReq, commit bool) error {
	// check required
	if len(r.RCTypeID) <= 0 {
		return errors.New("Missing `rcTypeId` field value")
	}
	if r.Quota <= 0 {
		return errors.New("Invalid `quota` value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
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

	// commit changes
	if g.EnableRaft && commit {
		cmd, err := json.Marshal(r)
		if err != nil {
			return err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Raft apply error, %v", err)
		}
		if rp := future.Response(); rp != nil {
			return fmt.Errorf("Raft apply response error, %v", rp)
		}
	}

	glog.V(1).Infof("Regist quota for rcType '%s' OK, total:%d", rcTypeIdHex, r.Quota)

	return nil
}

// doBorrow 借一个资源返回
func (l *limiterV1) doBorrow(r *APIBorrowReq, commit bool) (*APIBorrowResp, error) {
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

	// check leader
	if l.IsLeader() == false {
		return nil, ErrNotLeader
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

	// commit changes
	if g.EnableRaft && commit {
		cmd, err := json.Marshal(r)
		if err != nil {
			return nil, err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return nil, fmt.Errorf("Raft apply error, %v", err)
		}
		if frp := future.Response(); frp != nil {
			return nil, fmt.Errorf("Raft apply response error, %v", frp)
		}
	}

	//glog.V(1).Infof("Client[%s] borrow '%s' for %d seconds OK", r.CID.String(), rcId, r.Expire)

	return &APIBorrowResp{RCID: rcId}, nil
}

func (l *limiterV1) doReturn(r *APIReturnReq, commit bool) error {
	// check required
	if len(r.RCID) <= 0 {
		return errors.New("Missing `rcId` field value")
	}
	if len(r.ClientID) <= 0 {
		return errors.New("Missing `clientId` field value")
	}

	// check leader
	if l.IsLeader() == false {
		return ErrNotLeader
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

	// commit changes
	if g.EnableRaft && commit {
		cmd, err := json.Marshal(r)
		if err != nil {
			return err
		}

		future := l.raft.Apply(cmd, g.RaftTimeout)
		if err = future.Error(); err != nil {
			return fmt.Errorf("Raft apply error, %v", err)
		}
		if frp := future.Response(); frp != nil {
			return fmt.Errorf("Raft apply response error, %v", frp)
		}
	}

	//glog.V(1).Infof("Client[%s] return '%s' OK", r.CID.String(), r.RCID)

	return nil
}

func (l *limiterV1) doReturnAll(r *APIDisconnectReq, commit bool) error {
	//if len(r.CID) <= 0 {
	//	return errors.New("Missing `cId` field value")
	//}

	//// check leader
	//if l.IsLeader() == false {
	//	return ErrNotLeader
	//}

	//// try borrow
	//l.mutex.RLock()
	//defer l.mutex.Unlock()

	//tIdHex := r.TID.String()
	//m, ok := l.meta[tIdHex]
	//if !ok {
	//	return fmt.Errorf("ResourceType(%s) is not registed", tIdHex)
	//}
	//if err := m.ReturnAll(r.CID); err != nil {
	//	return err
	//}

	//// commit changes
	//if g.EnableRaft && commit {
	//	cmd, err := json.Marshal(r)
	//	if err != nil {
	//		return err
	//	}

	//	future := l.raft.Apply(cmd, g.RaftTimeout)
	//	if err = future.Error(); err != nil {
	//		return fmt.Errorf("Raft apply error, %v", err)
	//	}
	//	if rp := future.Response(); rp != nil {
	//		return fmt.Errorf("Raft apply response error, %v", rp)
	//	}
	//}

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
