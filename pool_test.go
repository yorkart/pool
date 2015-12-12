package pool

import (
	"fmt"
	"testing"
)

type ClientWraper struct {
}

func (p *ClientWraper) Close() {
	fmt.Println("close")
}

func (p *ClientWraper) Send() {
	fmt.Println("send")
}

func factory() (Conn, error) {
	return &ClientWraper{}, nil
}

func TestPool(t *testing.T) {

	conf := &Config{
		InitialCap:  5,
		MaxCap:      10,
		IdleTimeout: 10,
	}

	p, err := NewPool(conf, factory)
	if err != nil {
		fmt.Println("create pool error ", err)
	}

	for i:=0; i< 10; i++ {
		c,err := p.Checkout()
		if err != nil {
			fmt.Println("checkout error ", err)
		} else {
			c.(*ClientWraper).Send()
		}
	}

	p.Close()
}
