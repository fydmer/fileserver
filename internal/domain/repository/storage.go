package repository

import (
	"context"
	"time"
)

type StorageShardStatus int

const (
	StorageShardStatusOK = StorageShardStatus(iota)

	StorageShardStatusNew
	StorageShardStatusInProgress
	StorageShardStatusError
)

type StorageCreateShard struct {
	NodeId string
	Index  int
	Size   int64
}

type StorageCreateFileIn struct {
	Location string
	Shards   []*StorageCreateShard
}

type StorageCreateFileOut struct {
	Id string
}

type StorageSetShardStatusIn struct {
	FileId string
	NodeId string
	Index  int
	Status StorageShardStatus
}

type StorageSetShardStatusOut struct{}

type StorageGetFileIn struct {
	FileId string
}

type StorageShard struct {
	NodeId    string
	Index     int
	Size      int64
	CreatedAt time.Time
	Status    StorageShardStatus
}

type StorageGetFileOut struct {
	Id       string
	Location string
	Shards   []*StorageShard
}

type StorageGetFileByLocationIn struct {
	Location string
}

type StorageGetFileByLocationOut = StorageGetFileOut

type StorageDeleteFileIn struct {
	Id string
}

type StorageDeleteFileOut struct{}

type Storage interface {
	CreateFile(ctx context.Context, in *StorageCreateFileIn) (*StorageCreateFileOut, error)
	SetShardStatus(ctx context.Context, in *StorageSetShardStatusIn) (*StorageSetShardStatusOut, error)
	GetFile(ctx context.Context, in *StorageGetFileIn) (*StorageGetFileOut, error)
	GetFileByLocation(ctx context.Context, in *StorageGetFileByLocationIn) (*StorageGetFileByLocationOut, error)
	DeleteFile(ctx context.Context, in *StorageDeleteFileIn) (*StorageDeleteFileOut, error)
}
