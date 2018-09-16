package ratelimiter

import (
	"errors"
	"fmt"
	"sync"
)

type SessionMgrInterface interface {
	// Open 打开会话管理器
	Open(conf *SessionMgrConfig, transfer TransferInterface) error

	// GetSession 获取一个session实例
	GetSession() SessionInterface

	// Stage 暂存会话
	Stage(s *SessionInterface) error

	// Unstage 取消暂存
	Unstage(sessionId uint32) error
}

type SessionMgrConfig struct {
	// MaxCallbackBacklog 最大回调积压
	MaxCallbackBacklog int

	// ReadWorkers 读取响应的协程数
	ReadWorkers int
}

func DefaultSessionMgr() *SessionMgrInterface {
	return newSessionMgrV1()
}

func DefaultSessionMgrConfig() *SessionMgrConfig {
	return &SessionMgrConfig{
		MaxCallbackBacklog: 64,
		ReadWorkers:        8,
	}
}

type sessionMgrV1 struct {
	// conf 会话配置
	conf *SessionMgrConfig

	// sequence 全局自增的session序列
	sequence uint32

	// staged 暂存的会话
	staged map[uint32]SessionInterface

	// transfer 传输层实例
	transfer TransferInterface

	// readCh 读取传输层回调的响应
	readCh chan *TransferResponse

	// pool 会话对象复用池
	pool *sync.Pool

	// stopCh 控制关闭manager
	stopCh chan struct{}
}

func newSessionMgrV1() *SessionMgrInterface {
	return &sessionMgrV1{
		conf:     nil,
		sequence: 0,
		staged:   make(map[uint32]SessionInterface),
		transfer: nil,
		readCh:   make(chan *TransferResponse, 32),
		pool: sync.Pool{
			New: DefaultSession(),
		},
		stopCh: make(chan struct{}),
	}
}

func (mgr *sessionMgrV1) Open(conf *SessionMgrConfig, transfer TransferInterface) error {
	var err error

	// 校验配置
	if err = ValidateSessionMgrConfig(conf); err != nil {
		return err
	}

	//transfer.SetCallback(mgr.readCh)

	return err
}

func (mgr *sessionMgrV1) GetSession() SessionInterface {
	// TODO
	return nil
}

func (mgr *sessionMgrV1) Stage(s SessionInterface) error {
	// TODO
	return nil
}

func (mgr *sessionMgrV1) Unstage(sessionId uint32) error {
	// TODO:
	return nil
}

func (mgr *sessionMgrV1) receiveResponseLoop() {
	for {
		select {
		case rsp := <-mgr.readCh:
			// TODO
			// 查找staged，然后s.Done(rsp)
		case <-mgr.stopCh:
			return
		}
	}
}

func ValidateSessionMgrConfig(conf *SessionMgrConfig) error {
	if conf == nil {
		return errors.New("Config is nil")
	}
	if conf.MaxCallbackBacklog <= 0 {
		return fmt.Errorf("'SessionMgrConfig.MaxCallbackBacklog' should be greater than zero")
	}
	if conf.ReadWorkers <= 0 {
		return fmt.Errorf("'SessionMgrConfig.ReadWorkers' should be greater than zero")
	}
	return nil
}
