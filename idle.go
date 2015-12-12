package pool

import "time"

type Idel interface {
	CanRelease(cap,stock int, lastTime time.Time) bool
}

type FooIdel struct {

}

func (i *FooIdel) CanRelease(cap,stock int, lastTime int64) bool {

}