package pool

import (
	"fmt"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := make(chan bool, 10)
	in(c)

	time.Sleep(2 * time.Second)

	for i := 0; i < 10; i++ {
		fmt.Println(<-c)
	}
}

func in(c chan bool) {
	for i := 0; i < 20; i++ {
		fmt.Println("begin write ", i)
		select {
		case c <- true:
			fmt.Println("end write ", len(c))
		default:
			fmt.Println("chan is full")
		}
	}
}
