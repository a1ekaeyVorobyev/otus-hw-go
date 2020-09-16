package main

import (
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
	GetAdress() string
	Receive() error
	Send() error
}

type client struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	con     net.Conn
}

// NewTelnetClient create new telnetClient.
func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	// Place your code here

	return &client{address: address, timeout: timeout, in: in, out: out}
}

// Place your code here
// P.S. Author's solution takes no more than 50 lines.
func (c *client) Connect() (err error) {
	c.con, err = net.DialTimeout("tcp", c.address, c.timeout)

	return
}

func (c *client) GetAdress() string {
	return c.address
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
