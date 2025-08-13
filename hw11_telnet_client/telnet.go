//nolint:revive,nolintlint
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
		out:     out,
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
	if c.conn == nil || c.in == nil {
		return fmt.Errorf("nil parameter received: connection=%v, input_stream=%v", c.conn == nil, c.in == nil)
	}
	c.mu.RUnlock()

	data, err := c.readOut(c.in)
	if err != nil {
		return err
	}
	// CTRL+D case.
	if len(data) == 0 {
		return ErrEOT
	}

	err = c.writeOut(c.conn, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Receive() error {
	c.mu.RLock()
	if c.conn == nil || c.out == nil {
		return fmt.Errorf("nil parameter received: connection=%v, output_stream=%v", c.conn == nil, c.out == nil)
	}
	c.mu.RUnlock()

	data, err := c.readOut(c.conn)
	if err != nil {
		return err
	}

	err = c.writeOut(c.out, data)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) readOut(r io.Reader) ([]byte, error) {
	reader := bufio.NewReader(r)
	var res []byte
	for {
		line, err := reader.ReadBytes('\n')
		res = append(res, line...)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if reader.Buffered() == 0 {
					return nil, ErrEOT // No more data to read.
				}
				continue // More data in buffer, trying again.
			}
			if errors.Is(err, net.ErrClosed) {
				return nil, fmt.Errorf("%w: %w", ErrConnClosed, err)
			}
			return nil, fmt.Errorf("reading failed: %w", err)
		}
		if reader.Buffered() == 0 {
			break // Not expecting any more data here.
		}
	}
	return res, nil
}

func (c *Client) writeOut(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	if err != nil {
		if errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("%w: %w", ErrConnClosed, err)
		}
		return fmt.Errorf("unable to write out the data: %w", err)
	}
	return nil
}
