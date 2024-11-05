package controller

import (
	"context"
	"time"

	"github.com/fydmer/fileserver/internal/domain/repository"
)

func calculateFileParts(fileSize int64, partsCount int) []int64 {
	partSize := fileSize / int64(partsCount)
	lastPartSize := fileSize % int64(partsCount)
	var parts []int64

	if lastPartSize == 0 {
		for i := 0; i < partsCount; i++ {
			parts = append(parts, partSize)
		}
	} else {
		for i := 0; i < partsCount-1; i++ {
			parts = append(parts, partSize)
		}
		parts = append(parts, partSize+lastPartSize)
	}

	return parts
}

type shardStatusUpdater struct {
	storage        repository.Storage
	fileId, nodeId string
	index          int
}

func newShardStatusUpdater(storage repository.Storage, fileId, nodeId string, index int) *shardStatusUpdater {
	return &shardStatusUpdater{
		storage: storage,
		fileId:  fileId,
		nodeId:  nodeId,
		index:   index,
	}
}

func (x *shardStatusUpdater) toSetShardStatusIn() *repository.StorageSetShardStatusIn {
	return &repository.StorageSetShardStatusIn{
		FileId: x.fileId,
		NodeId: x.nodeId,
		Index:  x.index,
	}
}

func (x *shardStatusUpdater) setInProgress(ctx context.Context) error {
	req := x.toSetShardStatusIn()
	req.Status = repository.StorageShardStatusInProgress
	_, err := x.storage.SetShardStatus(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (x *shardStatusUpdater) setError(ctx context.Context) error {
	req := x.toSetShardStatusIn()
	req.Status = repository.StorageShardStatusError
	_, err := x.storage.SetShardStatus(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (x *shardStatusUpdater) setOK(ctx context.Context) error {
	req := x.toSetShardStatusIn()
	req.Status = repository.StorageShardStatusOK
	_, err := x.storage.SetShardStatus(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func errWithRollback(err error, rollback []func(ctx context.Context)) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	for _, fn := range rollback {
		fn(ctx)
	}
	return err
}
