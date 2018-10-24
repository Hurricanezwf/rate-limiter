package tcp

import (
	"errors"
	"sync"
	"time"
)

type DispatcherConfig struct {
	QueueSize       int
	WorkerNum       int
	PushBackTimeout time.Duration
}

type Event struct {
	Conn   *Connection
	Seq    uint32
	Action byte
	Msg    []byte
}

type EventRouter func(e *Event)

type EventDispatcher struct {
	//
	conf *DispatcherConfig

	//
	queue chan *Event

	//
	router EventRouter

	//
	wg *sync.WaitGroup

	//
	stopC chan struct{}
}

func (mgr *EventDispatcher) Open(conf *DispatcherConfig, router EventRouter) (err error) {
	if err = ValidateDispatcherConf(conf); err != nil {
		return err
	}
	if router == nil {
		return errors.New("EventRouter is nil")
	}

	mgr.conf = conf
	mgr.queue = make(chan *Event, conf.QueueSize)
	mgr.router = router
	mgr.wg = &sync.WaitGroup{}
	mgr.stopC = make(chan struct{})

	mgr.wg.Add(conf.WorkerNum)
	for i := 0; i < conf.WorkerNum; i++ {
		go mgr.handleEventsLoop()
	}

	return nil
}

func (mgr *EventDispatcher) Close() error {
	close(mgr.stopC)
	mgr.wg.Wait()
	return nil
}

func (mgr *EventDispatcher) PushBack(e *Event) error {
	select {
	case mgr.queue <- e:
		return nil
	case <-time.After(mgr.conf.PushBackTimeout):
		return errors.New("Too many events")
	}
}

func (mgr *EventDispatcher) handleEventsLoop() {
	for {
		select {
		case <-mgr.stopC:
			return
		case e := <-mgr.queue:
			if mgr.router != nil {
				mgr.router(e)
			}
		}
	}
}

func ValidateDispatcherConf(conf *DispatcherConfig) error {
	if conf == nil {
		return errors.New("Missing `DispatcherConfig`")
	}
	if conf.QueueSize <= 0 {
		return errors.New("DispatcherConfig.QueueSize too small")
	}
	if conf.WorkerNum <= 0 {
		return errors.New("DispatcherConfig.WorkerNum too little")
	}
	if conf.PushBackTimeout <= time.Second {
		return errors.New("Invalid DispatcherConfig.PushBackTimeout")
	}
	return nil
}
