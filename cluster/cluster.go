package cluster

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/Hurricanezwf/rate-limiter/g"
	"github.com/Hurricanezwf/rate-limiter/meta"
	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/logging/glog"
	"github.com/gogo/protobuf/proto"
	raftlib "github.com/hashicorp/raft"
)

func init() {
	RegistBuilder("v2", newClusterV2)
}

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
		m:           meta.New("v2"),
		raftTimeout: time.Duration(g.Config.Raft.Timeout) * time.Millisecond,
		stopC:       make(chan struct{}),
	}
}

func (c *clusterV2) Open() error {
	// TODO:
	return nil
}

func (c *clusterV2) IsLeader() bool {
	if g.Config.Raft.Enable == false {
		return true
	}
	return c.raft.State() == raftlib.Leader
}

func (c *clusterV2) LeaderHTTPAddr() string {
	c.gLock.RLock()
	defer c.gLock.RUnlock()
	return c.leaderHTTPAddr
}

// Apply 主要接收从其他结点过来的已提交的操作日志，然后应用到本结点
func (c *clusterV2) Apply(log *raftlib.Log) interface{} {
	glog.V(4).Infof("Apply Log: %s", string(log.Data))

	switch log.Type {
	case raftlib.LogCommand:
		{
			// 解析远端传过来的命令日志
			//var r Request
			//if err := json.Unmarshal(log.Data, &r); err != nil {
			//	return fmt.Errorf("Bad request format for raft command, %v", err)
			//}

			// 尝试执行命令
			//return l.switchDo(&r)
			// TODO:
			return nil
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
	// TODO
	return nil, nil

	//l.mutex.RLock()
	//defer l.mutex.RUnlock()

	//start := time.Now()
	//glog.V(1).Info("Create snapshot starting...")

	//// 编码元数据
	//b, err := proto.Marshal(l.meta)
	//if err != nil {
	//	glog.Warning(err.Error())
	//	return nil, err
	//}

	//// 编码时间戳
	//ts := make([]byte, 8)
	//binary.BigEndian.PutUint64(ts, uint64(start.Unix()))

	//// 构造快照
	//buf := bytes.NewBuffer(make([]byte, 0, 10+len(b)))
	//buf.WriteByte(MagicNumber)
	//buf.WriteByte(ProtocolVersion)
	//buf.Write(ts)
	//buf.Write(b)

	//glog.V(1).Infof("Create snapshot finished, elapse:%v, totalSize:%d", time.Since(start), buf.Len())

	//return limiter.NewLimiterSnapshot(buf), nil
}

// Restore 从快照中恢复LimiterFSM
func (c *clusterV2) Restore(rc io.ReadCloser) error {
	// TODO
	return nil

	//defer rc.Close()

	//glog.V(1).Info("Restore snapshot starting...")

	//start := time.Now()
	//reader := bufio.NewReader(rc)

	//// 读取识别魔数
	//magicNumber, err := reader.ReadByte()
	//if err != nil {
	//	glog.Warningf("Read magic number failed, %v", err)
	//	return err
	//}
	//if magicNumber != MagicNumber {
	//	err = fmt.Errorf("Unknown magic number %#x", magicNumber)
	//	glog.Warning(err.Error())
	//	return err
	//}

	//// 读取识别协议版本
	//protocolVersion, err := reader.ReadByte()
	//if err != nil {
	//	glog.Warningf("Read protocol version failed, %v", err)
	//	return err
	//}
	//if protocolVersion != ProtocolVersion {
	//	err = fmt.Errorf("ProtocolVersion(%#x) is not match with %#x", protocolVersion, ProtocolVersion)
	//	glog.Warning(err.Error())
	//	return err
	//}

	//// 读取快照时间戳
	//ts := make([]byte, 8)
	//n, err := reader.Read(ts)
	//if err != nil {
	//	glog.Warningf("Read timestamp failed, %v", err)
	//	return err
	//}
	//if n != len(ts) {
	//	err = errors.New("Read timestamp failed, missing data")
	//	glog.Warning(err.Error())
	//	return err
	//}
	//tsInt64 := binary.BigEndian.Uint64(ts)
	//glog.V(1).Infof("Restore from snapshot created at %d", tsInt64)

	//// 读取元数据并解析
	//buf := bytes.NewBuffer(make([]byte, 0, 10240))
	//bodyBytesCount, err := buf.ReadFrom(reader)
	//if err != nil && err != io.EOF {
	//	glog.Warningf("Read meta data failed, %v", err)
	//	return err
	//}

	//meta := types.NewMap()
	//if err := proto.Unmarshal(buf.Bytes(), meta); err != nil {
	//	glog.Warningf("Decode meta data failed, %v", err)
	//	return err
	//}

	//// 替换元数据
	//l.mutex.Lock()
	//l.meta = meta
	//l.mutex.Unlock()

	//glog.V(1).Infof("Restore snapshot finished, elapse:%v, totalBytes:%d", time.Since(start), 10+bodyBytesCount)

	//return nil
}

// RegistQuota 注册资源配额
func (c *clusterV2) RegistQuota(r *APIRegistQuotaReq) *APIRegistQuotaResp {
	rp := &APIRegistQuotaResp{Code: 500}
	cmd, err := encodeCMD(ActionRegistQuota, r)
	if err != nil {
		rp.Msg = fmt.Sprintf("Encode cmd failed, %v", err)
		return rp
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

	future := c.raft.Apply(cmd, c.raftTimeout)
	if err = future.Error(); err != nil {
		rp.Msg = err.Error()
		return rp
	}
	return future.Response().(*APIReturnAllResp)
}

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

func resolveCMD(b []byte) (cmd byte, args []byte) {
	if len(b) < 1 {
		return 0, nil
	}
	return b[0], b[1:]
}

func resolveArgs(args []byte, v proto.Message) error {
	return proto.Unmarshal(args, v)
}
