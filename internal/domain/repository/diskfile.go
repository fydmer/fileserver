package repository

import (
	"context"
	"io"
)

type DiskfileWriteIn struct {
	Name   string
	Source io.Reader
}

type DiskfileWriteOut struct {
	Written int64
}

type DiskfileReadIn struct {
	Name        string
	Destination io.Writer
}

type DiskfileReadOut struct {
	Written int64
}

type DiskfileRemoveIn struct {
	Name string
}

type DiskfileRemoveOut struct{}

type Diskfile interface {
	Write(ctx context.Context, in *DiskfileWriteIn) (*DiskfileWriteOut, error)
	Read(ctx context.Context, in *DiskfileReadIn) (*DiskfileReadOut, error)
	Remove(ctx context.Context, in *DiskfileRemoveIn) (*DiskfileRemoveOut, error)
}
