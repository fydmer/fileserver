package node

import (
	"context"
	"errors"
	"os"

	"github.com/fydmer/fileserver/internal/domain/repository"
	"github.com/fydmer/fileserver/internal/domain/service"
)

type Node struct {
	diskfile repository.Diskfile
}

func NewNode(diskfile repository.Diskfile) *Node {
	return &Node{
		diskfile: diskfile,
	}
}

func (x *Node) SaveFile(ctx context.Context, in *service.NodeSaveFileIn) (*service.NodeSaveFileOut, error) {
	write, err := x.diskfile.Write(ctx, &repository.DiskfileWriteIn{
		Name:   in.Name,
		Source: in.DataReader,
	})
	if err != nil {
		_, _ = x.diskfile.Remove(ctx, &repository.DiskfileRemoveIn{Name: in.Name})
		return nil, err
	}
	return &service.NodeSaveFileOut{
		Written: write.Written,
	}, nil
}

func (x *Node) GetFile(ctx context.Context, in *service.NodeGetFileIn) (*service.NodeGetFileOut, error) {
	read, err := x.diskfile.Read(ctx, &repository.DiskfileReadIn{
		Name:        in.Name,
		Destination: in.DataWriter,
	})
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}
	return &service.NodeGetFileOut{
		Written: read.Written,
	}, nil
}

func (x *Node) DeleteFile(ctx context.Context, in *service.NodeDeleteFileIn) (*service.NodeDeleteFileOut, error) {
	if _, err := x.diskfile.Remove(ctx, &repository.DiskfileRemoveIn{Name: in.Name}); err != nil {
		return nil, err
	}
	return &service.NodeDeleteFileOut{}, nil
}
