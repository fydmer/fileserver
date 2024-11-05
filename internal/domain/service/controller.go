package service

import (
	"context"
	"io"
)

type ControllerJoinNodeIn struct {
	Addr string
}

type ControllerJoinNodeOut struct {
	Id string
}

type ControllerUploadFileIn struct {
	Location string
	Size     int64
	Content  io.Reader
}

type ControllerUploadFileOut struct {
	Id string
}

type ControllerSearchFileIn struct {
	Location string
}

type ControllerSearchFileOut struct {
	Id     string
	Size   int64
	Status int
}

type ControllerDownloadFileIn struct {
	Id      string
	Content io.Writer
}

type ControllerDownloadFileOut struct{}

type ControllerDeleteFileIn struct {
	Id string
}

type ControllerDeleteFileOut struct{}

type Controller interface {
	JoinNode(ctx context.Context, in *ControllerJoinNodeIn) (*ControllerJoinNodeOut, error)
	UploadFile(ctx context.Context, in *ControllerUploadFileIn) (*ControllerUploadFileOut, error)
	SearchFile(ctx context.Context, in *ControllerSearchFileIn) (*ControllerSearchFileOut, error)
	DownloadFile(ctx context.Context, in *ControllerDownloadFileIn) (*ControllerDownloadFileOut, error)
	DeleteFile(ctx context.Context, in *ControllerDeleteFileIn) (*ControllerDeleteFileOut, error)
}
