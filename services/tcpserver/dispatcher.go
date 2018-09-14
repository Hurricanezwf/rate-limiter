package tcpserver

import (
	"errors"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/proto"
)

type DispatcherConfig struct {
	QueueSize int
	Worker    int
}

type Event struct {
	Conn    *Connection
	Seq     uint32
	Action  byte
	Msg     []byte
	Limiter *limiter.Interface
}

type EventDispatcher struct {
	//
	conf *Config

	//
	queue chan *Event

	//
	wg *sync.WaitGroup

	//
	stopC chan struct{}
}

func (mgr *EventDispatcher) Open(conf *DispatcherConfig) (err error) {
	if err = ValidateConf(conf); err != nil {
		return err
	}

	conf = conf
	queue = make(chan *Event, conf.QueueSize)
	wg = &sync.WaitGroup{}
	stopC = make(chan struct{})

	wg.Add(conf.Worker)
	for i := 0; i < conf.Worker; i++ {
		go handleEventsLoop()
	}

	return nil
}

func (mgr *EventDispatcher) Close() error {
	close(mgr.stopC)
	mgr.wg.Wait()
}

func (mgr *EventDispatcher) PushBack(e *Event) error {
	select {
	case mgr.queue <- e:
		return nil
	}
	return errors.New("Too many events")
}

func (mgr *EventDispatcher) handleEventsLoop() {
	for {
		select {
		case <-mgr.stopC:
			return
		case e := <-mgr.queue:
			mgr.hanle(e)
		}
	}
}

func (mgr *EventDispatcher) handle(e *Event) {
	switch e.Action {
	case proto.ActionBorrow:
		e.Limiter.BorrowWith(e.Msg)
	case proto.ActionReturn:
		e.Limiter.ReturnWith(e.Msg)
	case proto.ActionReturnAll:
		e.Limiter.ReturnAllWith(e.Msg)
	case proto.ActionRegistQuota:
		e.Limiter.RegistQuotaWith(e.Msg)
	case proto.ActionDeleteQuota:
		e.Limiter.DeleteQuotaWith(e.Msg)
	case proto.ActionResourceList:
		e.Limiter.ResourceListWith(e.Msg)
	}
}

func ValidateDispatcherConf(conf *DispatcherConfig) error {
	if conf.QueueSize <= 0 {
		return errors.New("QueueSize too small")
	}
	if conf.Worker <= 0 {
		return errors.New("Worker count too little")
	}
	return nil
}
