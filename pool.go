package pool

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
	"fmt"
)

// ConnectionPool connection pool manager
type Pool struct {
	waitings  int32
	hits      int64

	available bool

	cap    int
	maxCap int
	queue chan Conn

	factory Factory

	lock *sync.Mutex
}

// NewPool create ConnectionPool instance
func NewPool(initialCap, maxCap int, factory Factory) (*Pool,error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	p := &Pool{
		available: false,

		queue: make(chan Conn, maxCap),
		maxCap : maxCap,

		factory:   factory,

		lock: &sync.Mutex{},
	}

	for i := 0; i < initialCap; i++ {
		_, err := p.create()
		if err != nil {
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
	}

	go p.idleCheck()

	return p, nil
}

// Create get a new client session
func (p *Pool) create() (Conn, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if len(p.queue) >= p.maxCap {
		return nil, errors.New("pool is full")
	}

	c, err := p.factory()
	if err != nil {
		p.Close()
		return nil, fmt.Errorf("create connection error: %s", err)
	}

	if err = p.add(c);err != nil {
		return nil ,err
	}
	p.cap ++
	return c, nil
}

func (p *Pool) add(c Conn) error{
	select {
	case p.queue<-c :
		return nil
	default :
		c.Close()
		return errors.New("pool is full")
	}
}

func (p *Pool) get() (Conn, error){
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

	c, err := p.get();
	if  err != nil {
		return nil,err
	}

	atomic.AddInt32(&p.hits, 1)
	return c, nil
}

// Checkin give back client session
func (p *Pool) Checkin(c Conn) error{
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

	for i:=0; i< p.maxCap; i++ {
		c := <- p.queue
		c.Close()
	}

	close(p.queue)
	p.queue = nil
}

// Destroy close a client session
func (p *Pool) destroy(c Conn) {
	c.Close()
	p.cap --
}

func (p *Pool) idleCheck() {
	for {
		if !p.available {
			break
		}
		p.tryRelease()
		time.Sleep(1 * time.Second)
	}
}

func (p *Pool) tryRelease() {
	if !p.canRelease() {
		return
	}

	if client, err := p.Checkout(); err == nil && client != nil {
		p.release(client)
	} else {
		p.Checkin(client)
	}
}

func (p *Pool) canRelease() bool {
	cap := int(p.cap)
	len := len(p.queue)

	if cap < 3 {
		return false
	}

	if len > int(float32(cap)*0.5) {
		return true
	}
	return false
}

func (p *Pool) release(c Conn) bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	if !p.available {
		return false
	}

	p.destroy(c)


	return true
}
