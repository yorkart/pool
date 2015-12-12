package pool

import "time"

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
	ts := p.hitTS
	if time.Now().Sub(time.Unix(ts,0)).Seconds() > float64(p.conf.IdleTimeout) {
		return true
	}

	cap := int(p.cap)
	stock := len(p.queue)

	if cap < 3 {
		return false
	}

	if stock > int(float32(cap)*0.5) {
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
