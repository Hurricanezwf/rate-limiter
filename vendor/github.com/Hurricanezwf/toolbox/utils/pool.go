package utils

import (
	"bytes"
)

var (
	BytesBuffer = NewBuffersPool(64)
)

// BuffersPool is *bytes.Buffer's pool which limit the max size of pool.
type BuffersPool struct {
	pool chan *bytes.Buffer
}

func NewBuffersPool(max int) *BuffersPool {
	if max <= 0 {
		return nil
	}
	return &BuffersPool{
		pool: make(chan *bytes.Buffer, max),
	}
}

func (p *BuffersPool) Get(minCap int) *bytes.Buffer {
	var b *bytes.Buffer
	select {
	case b = <-p.pool:
		if b.Cap() < minCap {
			b.Grow(minCap - b.Cap())
		}
		b.Reset()
	default:
		b = bytes.NewBuffer(nil)
		b.Grow(minCap)
	}
	return b
}

func (p *BuffersPool) Put(b *bytes.Buffer) {
	if b == nil {
		return
	}
	select {
	case p.pool <- b:
	}
}

// LimitPool; 带缓冲大小的buffer
type LimitPool struct {
	limit int
	p     chan interface{}
}

func NewLimitPool(limit int) *LimitPool {
	if limit <= 0 {
		panic("BufferedPool: limit <= 0")
	}

	return &LimitPool{
		limit: limit,
		p:     make(chan interface{}, limit),
	}
}

// Fill: 将所有元素填充为nil
func (l *LimitPool) Fill() *LimitPool {
	for i := 0; i < l.limit; i++ {
		l.p <- nil
	}
	return l
}

func (l *LimitPool) Put(v interface{}) {
	l.p <- v
}

func (l *LimitPool) Get() interface{} {
	return <-l.p
}

func (l *LimitPool) Destroy() {
	l.limit = 0
	close(l.p)
}

func (l *LimitPool) Chan() chan interface{} {
	return l.p
}

// SpeedLimiter
type SpeedLimiter struct {
	block bool
	limit int
	p     chan interface{}
}

func NewSpeedLimiter(limit int, block bool) *SpeedLimiter {
	if limit <= 0 {
		panic("BufferedPool: limit <= 0")
	}

	return &SpeedLimiter{
		block: block,
		limit: limit,
		p:     make(chan interface{}, limit),
	}
}

func (l *SpeedLimiter) Authorize() bool {
	if l.block == false {
		select {
		case l.p <- nil:
			return true
		}
	} else {
		l.p <- nil
		return true
	}
	return false
}

func (l *SpeedLimiter) Revoke() bool {
	if l.block == false {
		select {
		case <-l.p:
			return true
		}
	} else {
		<-l.p
		return true
	}
	return false
}

func (l *SpeedLimiter) Destroy() {
	l.limit = 0
	close(l.p)
}
