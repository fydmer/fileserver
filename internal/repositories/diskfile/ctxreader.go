package diskfile

import (
	"context"
	"io"
)

type ctxReader struct {
	ctx       context.Context
	chunkSize int
	reader    io.Reader
}

func (x *ctxReader) Read(p []byte) (int, error) {
	select {
	case <-x.ctx.Done():
		return 0, x.ctx.Err()
	default:
	}

	limitedReader := io.LimitReader(x.reader, int64(x.chunkSize))

	return limitedReader.Read(p)
}
