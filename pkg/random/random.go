package random

import (
	"context"
	"crypto/rand"
	"io"
)

func WriteRandomData(ctx context.Context, dst io.Writer, size int64) error {
	buf := make([]byte, 4096)
	var written int64
	for written < size {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		wr := buf
		if remaining := size - written; remaining < int64(len(buf)) {
			wr = buf[:remaining]
		}
		if _, err := rand.Read(wr); err != nil {
			return err
		}
		if _, err := dst.Write(wr); err != nil {
			return err
		}
		written += int64(len(wr))
	}

	return nil
}
