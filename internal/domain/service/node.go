package service

import (
	"context"
	"io"
)

type NodeSaveFileIn struct {
	Name       string
	DataReader io.Reader
}

type NodeSaveFileOut struct {
	Written int64
}

type NodeGetFileIn struct {
	Name       string
	DataWriter io.Writer
}

type NodeGetFileOut struct {
	Written int64
}

type NodeDeleteFileIn struct {
	Name string
}

type NodeDeleteFileOut struct{}

type Node interface {
	SaveFile(ctx context.Context, in *NodeSaveFileIn) (*NodeSaveFileOut, error)
	GetFile(ctx context.Context, in *NodeGetFileIn) (*NodeGetFileOut, error)
	DeleteFile(ctx context.Context, in *NodeDeleteFileIn) (*NodeDeleteFileOut, error)
}
