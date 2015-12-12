package pool

import "errors"

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

type Conn interface {
	Close()
}

// Factory is a function to create new connections.
type Factory func() (Conn, error)
