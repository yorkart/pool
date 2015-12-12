package pool

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionPool connection pool manager
type Pool struct {
	waitings int32
	hits     int64

	available bool

	cap     int
	queue   chan Conn
	hitTS int64

	factory Factory
	conf    *Config

	lock *sync.Mutex
}

// NewPool create ConnectionPool instance
func NewPool(conf *Config, factory Factory) (*Pool, error) {
	confCopy, err := conf.copy()
	if err != nil {
		return nil, err
	}

	p := &Pool{
		queue: make(chan Conn, confCopy.MaxCap),

		factory: factory,
		conf:    confCopy,

		lock: &sync.Mutex{},
	}

	for i := 0; i < p.conf.InitialCap; i++ {
		_, err := p.create()
		if err != nil {
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
	}

	go p.idleCheck()
	p.available = true
	return p, nil
}

// Create get a new client session
func (p *Pool) create() (Conn, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.queue) >= p.conf.MaxCap {
		return nil, errors.New("pool is full")
	}

	c, err := p.factory()
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("create connection error: %s", err)
	}

	if err = p.add(c); err != nil {
		return nil, err
	}
	p.cap++
	return c, nil
}

func (p *Pool) add(c Conn) error {
	select {
	case p.queue <- c:
		return nil
	default:
		c.Close()
		return errors.New("pool is full")
	}
}

func (p *Pool) get() (Conn, error) {
	select {
	case c := <-p.queue:
		if c == nil {
			return nil, ErrClosed
		}
		return c, nil
	default:
		c, err := p.create()
		if err != nil {
			return nil, err
		}
		return c, nil
	}
}

// Checkout try get a client session
func (p *Pool) Checkout() (Conn, error) {
	if !p.available {
		return nil, ErrClosed
	}

	c, err := p.get()
	if err != nil {
		return nil, err
	}

	atomic.AddInt64(&p.hits, 1)
	atomic.StoreInt64(&p.hitTS, time.Now().Unix())

	return c, nil
}

// Checkin give back client session
func (p *Pool) Checkin(c Conn) error {
	if c == nil {
		return errors.New("connection is nil. rejecting")
	}

	return p.add(c)
}

// Close close connection pool
func (p *Pool) Close() {
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.available {
		return
	}
	p.available = false

	for i := 0; i < p.cap; i++ {
		c,err := p.Checkout()
		if err != nil {
			i--
			continue
		}
		c.Close()
	}

	close(p.queue)
	p.queue = nil
}

// Destroy close a client session
func (p *Pool) destroy(c Conn) {
	c.Close()
	p.cap--
}
