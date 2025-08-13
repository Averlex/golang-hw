//nolint:revive
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

var (
	ErrEOT        = errors.New("EOT signal received")
	ErrConnClosed = errors.New("connection is closed")
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

type Client struct {
	mu      sync.RWMutex
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &Client{
		address: address,
		timeout: timeout,
		in:      in,
		out:     bufio.NewWriter(out),
	}
}

func (c *Client) Connect() error {
	c.mu.RLock()
	address, timeout := c.address, c.timeout
	c.mu.RUnlock() // To avoid blocking while dialing with timeout.
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var err error
	switch {
	case c.conn != nil && c.in != nil:
		inErr := c.in.Close()
		err = c.conn.Close()
		if err == nil {
			err = inErr
		}
	case c.conn != nil:
		err = c.conn.Close()
	case c.in != nil:
		err = c.in.Close()
	default:
		return nil
	}
	c.conn = nil
	return err
}

func (c *Client) Send() error {
	c.mu.RLock()
	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}

	if c.in == nil {
		return fmt.Errorf("input stream is not initialized")
	}
	c.mu.RUnlock()

	reader := bufio.NewReader(c.in)
	line, err := reader.ReadString('\n')
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("read from input failed: %w", err)
		}
	}
	if len(line) == 0 {
		return ErrEOT
	}

	_, err = c.conn.Write([]byte(line))
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", ErrConnClosed, err)
		}
		return fmt.Errorf("unable to send data: %w", err)
	}
	return nil
}

func (c *Client) Receive() error {
	c.mu.RLock()

	if c.out == nil {
		return fmt.Errorf("output stream is not initialized")
	}

	if c.conn == nil {
		return fmt.Errorf("connection is not initialized")
	}
	c.mu.RUnlock()

	r := bufio.NewReader(c.conn)
	line, err := r.ReadString('\n')
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", ErrConnClosed, err)
		}
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("read from connection failed: %w", err)
		}
		// Reading EOF is ok here.
	}

	_, err = c.out.Write([]byte(line))
	if err != nil {
		return fmt.Errorf("unable to write out the data: %w", err)
	}

	// To prevent stdout from being buffered indefinitely -> removing the delay in output.
	if bw, ok := c.out.(*bufio.Writer); ok {
		err = bw.Flush()
		if err != nil {
			return fmt.Errorf("unable to flush output: %w", err)
		}
	}

	return nil
}
