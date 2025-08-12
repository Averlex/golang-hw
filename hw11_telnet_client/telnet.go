//nolint:revive
package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"
)

var ErrEOT = errors.New("EOT signal received")

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type Client struct {
	address   string
	timeout   time.Duration
	in        io.ReadCloser
	out       io.Writer
	mu        sync.Mutex
	conn      net.Conn
	logWriter io.Writer
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &Client{
		address:   address,
		timeout:   timeout,
		in:        in,
		out:       out,
		logWriter: os.Stderr,
	}
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	var err error
	c.conn, err = net.DialTimeout("tcp", c.address, c.timeout)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	_, _ = fmt.Fprintf(c.logWriter, "...Connected to %s with timeout %v\n", c.address, max(0, c.timeout))
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.conn.Close()
	// Prioritizing conn.Close() of in.Close().
	if c.in != nil {
		if err == nil {
			err = c.in.Close()
		} else {
			c.in.Close()
		}
	}
	if err != nil {
		return fmt.Errorf("connection close failed: %w", err)
	}
	_, _ = fmt.Fprintf(c.logWriter, "...Connection successfully closed\n")
	c.conn = nil
	return nil
}

func (c *Client) Send() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}

	if c.in == nil {
		return fmt.Errorf("input stream is not initialized")
	}

	buffer, err := io.ReadAll(c.in)
	if err != nil {
		return fmt.Errorf("unexpected error occurred during reading the data: %w", err)
	}
	// EOF received -> io.ReadAll does not return an error on getting EOF.
	if len(buffer) == 0 {
		return ErrEOT
	}
	n, err := c.conn.Write(buffer)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", ErrEOT, err)
		}
		return fmt.Errorf("unable to send data: %w", err)
	}
	_, _ = fmt.Fprintf(c.logWriter, "...New message sent: len=%v bytes\n", n)
	return nil
}

func (c *Client) Receive() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.out == nil {
		return fmt.Errorf("output stream is not initialized")
	}

	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}

	buffer, err := io.ReadAll(c.conn)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", ErrEOT, err)
		}
		return fmt.Errorf("unexpected error occurred during reading the data: %w", err)
	}
	// EOF received -> io.ReadAll does not return an error on getting EOF.
	if len(buffer) == 0 {
		return ErrEOT
	}
	n, err := c.out.Write(buffer)
	if err != nil {
		return fmt.Errorf("unable to write out the data: %w", err)
	}
	_, _ = fmt.Fprintf(c.logWriter, "...New message received: len=%v bytes\n", n)
	return nil
}
