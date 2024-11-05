package diskfile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fydmer/fileserver/internal/domain/repository"
)

type Repository struct {
	rootDir string
}

const chunkSize = 1 * 1024 * 1024

func NewRepository(rootDir string) (*Repository, error) {
	st, err := os.Stat(rootDir)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll(rootDir, 0755); err != nil {
			return nil, err
		}
		st, err = os.Stat(rootDir)
	}
	if err != nil {
		return nil, err
	}

	if !st.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", rootDir)
	}

	if st.Mode()&0400 == 0 && st.Mode()&0200 == 0 {
		return nil, fmt.Errorf("permission denied")
	}

	return &Repository{
		rootDir: rootDir,
	}, nil
}

func (x *Repository) Write(ctx context.Context, in *repository.DiskfileWriteIn) (*repository.DiskfileWriteOut, error) {
	f, err := os.Create(filepath.Join(x.rootDir, in.Name))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := &ctxReader{ctx: ctx, chunkSize: chunkSize, reader: in.Source}

	written, err := io.Copy(f, r)
	if err != nil {
		return nil, err
	}

	return &repository.DiskfileWriteOut{
		Written: written,
	}, nil
}

func (x *Repository) Read(ctx context.Context, in *repository.DiskfileReadIn) (*repository.DiskfileReadOut, error) {
	f, err := os.Open(filepath.Join(x.rootDir, in.Name))
	if err != nil {
		return nil, err
	}

	defer f.Close()

	r := &ctxReader{ctx: ctx, chunkSize: chunkSize, reader: f}

	written, err := io.Copy(in.Destination, r)
	if err != nil {
		return nil, err
	}
	return &repository.DiskfileReadOut{
		Written: written,
	}, nil
}

func (x *Repository) Remove(_ context.Context, in *repository.DiskfileRemoveIn) (*repository.DiskfileRemoveOut, error) {
	err := os.Remove(filepath.Join(x.rootDir, in.Name))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return &repository.DiskfileRemoveOut{}, nil
}
