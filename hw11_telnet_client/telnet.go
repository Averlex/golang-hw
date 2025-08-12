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
	mu        sync.RWMutex
	conn      net.Conn
	logWriter io.Writer
	buffer    []byte
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
	_, _ = fmt.Fprintf(c.logWriter, "...Connected to %s with timeout %v", c.address, max(0, c.timeout))
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
	_, _ = fmt.Fprintf(c.logWriter, "...Connection successfully closed")
	c.conn = nil
	return nil
}

func (c *Client) Send() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.in == nil {
		return fmt.Errorf("input stream is not initialized")
	}

	if c.buffer != nil {
		return fmt.Errorf("internal buffer is not empty - call Receive() first")
	}

	c.buffer = make([]byte, 0)
	n, err := c.in.Read(c.buffer)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return ErrEOT
		}
		return fmt.Errorf("unexpected error occurred during sending: %w", err)
	}
	_, _ = fmt.Fprintf(c.logWriter, "...New message sent: len=%vb", n)
	return nil
}

func (c *Client) Receive() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.out == nil {
		return fmt.Errorf("output stream is not initialized")
	}

	if c.conn == nil {
		return fmt.Errorf("connection is closed")
	}

	if c.buffer == nil {
		return fmt.Errorf("internal buffer is empty - nothing to receive")
	}

	buffer := c.buffer
	c.buffer = nil

	n, err := c.out.Write(buffer)
	if err != nil {
		return fmt.Errorf("unexpected error occurred during receiving")
	}
	_, _ = fmt.Fprintf(c.logWriter, "...New message received: len=%vb", n)
	return nil
}
