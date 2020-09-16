package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"
)

var _ TelnetClient = (*client)(nil)

// TelnetClient description interface.
type TelnetClient interface {
	// Place your code here
	Connect() error
	Close() error
	GetAdress() string
	Receive() error
	Send() error
	Cancel()
	GetContext() context.Context
}

type client struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	con     net.Conn
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewTelnetClient create new telnetClient.
func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	// Place your code here
	ctx, cancel := context.WithCancel(context.Background())

	return &client{address: address, timeout: timeout, in: in, out: out, ctx: ctx, cancel: cancel}
}

// Place your code here
// P.S. Author's solution takes no more than 50 lines.
func (c *client) Connect() (err error) {
	c.con, err = net.DialTimeout("tcp", c.address, c.timeout)

	return
}

func (c *client) Close() error {
	if c.con == nil {
		return fmt.Errorf("tcp connection is nil")
	}

	return c.con.Close()
}

func (c *client) GetAdress() string {
	return c.address
}

func (c *client) Cancel() {
	c.cancel()
}

func (c *client) Send() (e error) {
	if c.con == nil {
		return fmt.Errorf("tcp connection is nil")
	}
	_, e = io.Copy(c.con, c.in)

	return
}

func (c *client) Receive() (err error) {
	if c.con == nil {
		return fmt.Errorf("tcp connection is nil")
	}
	_, err = io.Copy(c.out, c.con)

	return
}

func (c *client) GetContext() context.Context {
	return c.ctx
}
