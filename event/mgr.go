package event

import (
	"errors"
	"sync"

	"github.com/Hurricanezwf/rate-limiter/limiter"
	"github.com/Hurricanezwf/rate-limiter/services/tcpserver"
)

type Config struct {
	QueueSize int
	Worker    int
}

type Event struct {
	Action  byte
	Msg     []byte
	Conn    *tcpserver.Connection
	Limiter *limiter.Interface
}

type EventManager struct {
	//
	conf *Config

	//
	queue chan *Event

	//
	wg *sync.WaitGroup

	//
	stopC chan struct{}
}

func (mgr *EventManager) Open(conf *Config) (err error) {
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

func (mgr *EventManager) Close() error {
	close(mgr.stopC)
	mgr.wg.Wait()
}

func (mgr *EventManager) PushBack(e *Event) error {
	select {
	case mgr.queue <- e:
		return nil
	}
	return errors.New("Overflow event manager")
}

func (mgr *EventManager) handleEventsLoop() {
	for {
		select {
		case <-mgr.stopC:
			return
		case e := <-mgr.queue:
			mgr.hanle(e)
		}
	}
}

func (mgr *EventManager) handle(e *Event) {

}

func ValidateConf(conf *Config) error {
	if conf.QueueSize <= 0 {
		return errors.New("QueueSize too small")
	}
	if conf.Worker <= 0 {
		return errors.New("Worker count too little")
	}
	return nil
}
