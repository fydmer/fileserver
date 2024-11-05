package nodecli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	addr   string
	dialer net.Dialer
}

const (
	sendFileChunkSize = 1024 * 1024
)

func NewClient(ctx context.Context, addr string) (*Client, error) {
	c := &Client{
		addr: addr,
		dialer: net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	}

	return c, nil
}

func (c *Client) SaveFile(ctx context.Context, filename string, src io.Reader, size int64) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	r, w := bufio.NewReader(conn), bufio.NewWriter(conn)

	if _, err = w.WriteString(fmt.Sprintf("save_file:%d:%s\n", size, filename)); err != nil {
		return fmt.Errorf("failed to send header: %w", err)
	}

	lr := io.LimitReader(src, size)

	var writtenSum int64
	chunkBuf := make([]byte, sendFileChunkSize)
	for {
		n, err := lr.Read(chunkBuf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk: %w", err)
		}
		if n == 0 {
			break
		}

		written, err := w.Write(chunkBuf[:n])
		if err != nil {
			return fmt.Errorf("failed to send chunk: %w", err)
		}
		writtenSum += int64(written)
	}
	if err = w.Flush(); err != nil {
		return fmt.Errorf("failed to send bufferized data: %w", err)
	}

	nodeWrittenStr, err := r.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to recieve written size: %w", err)
	}

	nodeWritten, err := strconv.ParseInt(strings.TrimSpace(nodeWrittenStr), 10, 64)
	if err != nil {
		return fmt.Errorf("failed to parse written size: %w", err)
	}

	if nodeWritten != writtenSum {
		return fmt.Errorf("written size does not match file size")
	}

	return nil
}

func (c *Client) GetFile(ctx context.Context, filename string, dst io.Writer, size int64) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(fmt.Sprintf("get_file:%s\n", filename))); err != nil {
		return fmt.Errorf("failed to send header: %w", err)
	}

	lr := io.LimitReader(conn, size)

	written, err := io.Copy(dst, lr)
	if err != nil {
		return fmt.Errorf("failed to receive file data: %w", err)
	}

	if _, err = conn.Write([]byte(fmt.Sprintf("%d\n", written))); err != nil {
		return fmt.Errorf("failed to send written size info: %w", err)
	}

	return nil
}

func (c *Client) DeleteFile(ctx context.Context, filename string) error {
	conn, err := c.dialer.DialContext(ctx, "tcp", c.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to node: %w", err)
	}
	defer conn.Close()

	if _, err = conn.Write([]byte(fmt.Sprintf("delete_file:%s\n", filename))); err != nil {
		return fmt.Errorf("failed to send header: %w", err)
	}

	return nil
}
