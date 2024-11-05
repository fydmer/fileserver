package controller

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/fydmer/fileserver/internal/domain/repository"
	"github.com/fydmer/fileserver/internal/domain/service"
	"github.com/fydmer/fileserver/pkg/nodecli"
)

type nodeClients struct {
	mu   sync.Mutex
	dict map[string]*nodecli.Client
}

type Controller struct {
	countFileParts int
	infra          repository.Infra
	storage        repository.Storage
	nodeClients    nodeClients
}

const defaultCountFileParts = 6

func NewController(infra repository.Infra, storage repository.Storage) *Controller {
	return &Controller{
		countFileParts: defaultCountFileParts,
		infra:          infra,
		storage:        storage,
		nodeClients: nodeClients{
			mu:   sync.Mutex{},
			dict: make(map[string]*nodecli.Client),
		},
	}
}

func (x *Controller) JoinNode(ctx context.Context, in *service.ControllerJoinNodeIn) (*service.ControllerJoinNodeOut, error) {
	createNode, err := x.infra.CreateNode(ctx, &repository.InfraCreateNodeIn{
		Addr: in.Addr,
	})
	if err != nil {
		return nil, err
	}
	return &service.ControllerJoinNodeOut{
		Id: createNode.Node.Id,
	}, nil
}

func (x *Controller) getNodeClient(ctx context.Context, node *repository.InfraNode) (*nodecli.Client, error) {
	x.nodeClients.mu.Lock()
	defer x.nodeClients.mu.Unlock()

	cli, ok := x.nodeClients.dict[node.Id]
	if ok {
		return cli, nil
	}

	cli, err := nodecli.NewClient(ctx, node.Addr)
	if err != nil {
		return nil, err
	}

	x.nodeClients.dict[node.Id] = cli
	return cli, nil
}

func (x *Controller) UploadFile(ctx context.Context, in *service.ControllerUploadFileIn) (*service.ControllerUploadFileOut, error) {
	parts := calculateFileParts(in.Size, x.countFileParts)

	freerNodes, err := x.infra.GetFreerNodes(ctx, &repository.InfraGetFreerNodesIn{
		Count: x.countFileParts,
	})
	if err != nil {
		return nil, err
	}
	if len(parts) != len(freerNodes.Nodes) {
		return nil, errors.New("wrong number of parts or number of nodes")
	}

	slices.Reverse(freerNodes.Nodes)

	var shards []*repository.StorageCreateShard
	for index, node := range freerNodes.Nodes {
		shards = append(shards, &repository.StorageCreateShard{
			NodeId: node.Id,
			Index:  index,
			Size:   parts[index],
		})
	}

	file, err := x.storage.CreateFile(ctx, &repository.StorageCreateFileIn{
		Location: in.Location,
		Shards:   shards,
	})
	if err != nil {
		return nil, err
	}

	var rollback []func(ctx context.Context)
	rollback = append(rollback, func(ctx context.Context) {
		if _, err := x.storage.DeleteFile(ctx, &repository.StorageDeleteFileIn{
			Id: file.Id,
		}); err != nil {
			slog.Error("error to clean wrong file record", slog.String("error", err.Error()))
		}
	})

	for index, size := range parts {
		nodeClient, err := x.getNodeClient(ctx, freerNodes.Nodes[index])
		if err != nil {
			return nil, errWithRollback(err, rollback)
		}

		st := newShardStatusUpdater(x.storage, file.Id, freerNodes.Nodes[index].Id, index)
		if err = st.setInProgress(ctx); err != nil {
			return nil, errWithRollback(err, rollback)
		}

		filename := fmt.Sprintf("%s.%d", file.Id, index)

		rollback = append(rollback, func(ctx context.Context) {
			if err := nodeClient.DeleteFile(ctx, filename); err != nil {
				slog.Error("error to clean wrong file data", slog.String("error", err.Error()))
			}
		})

		if err = nodeClient.SaveFile(ctx, filename, in.Content, size); err != nil {
			return nil, errors.Join(errWithRollback(err, rollback), st.setError(ctx))
		}

		if err = st.setOK(ctx); err != nil {
			return nil, errWithRollback(err, rollback)
		}
	}

	return &service.ControllerUploadFileOut{
		Id: file.Id,
	}, nil
}

func (x *Controller) SearchFile(ctx context.Context, in *service.ControllerSearchFileIn) (*service.ControllerSearchFileOut, error) {
	file, err := x.storage.GetFileByLocation(ctx, &repository.StorageGetFileByLocationIn{
		Location: in.Location,
	})
	if err != nil {
		return nil, err
	}

	var size int64
	status := repository.StorageShardStatusOK

	for _, shard := range file.Shards {
		size += shard.Size
		if shard.Status > status {
			status = shard.Status
		}
	}

	return &service.ControllerSearchFileOut{
		Id:     file.Id,
		Size:   size,
		Status: int(status),
	}, nil
}

func (x *Controller) DownloadFile(ctx context.Context, in *service.ControllerDownloadFileIn) (*service.ControllerDownloadFileOut, error) {
	file, err := x.storage.GetFile(ctx, &repository.StorageGetFileIn{
		FileId: in.Id,
	})
	if err != nil {
		return nil, err
	}

	slices.SortFunc(file.Shards, func(a, b *repository.StorageShard) int {
		return cmp.Compare(a.Index, b.Index)
	})

	for _, shard := range file.Shards {
		getNode, err := x.infra.GetNode(ctx, &repository.InfraGetNodeIn{
			Id: shard.NodeId,
		})
		if err != nil {
			return nil, err
		}

		cli, err := x.getNodeClient(ctx, getNode.Node)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s.%d", in.Id, shard.Index)

		if err = cli.GetFile(ctx, filename, in.Content, shard.Size); err != nil {
			return nil, err
		}
	}

	return &service.ControllerDownloadFileOut{}, nil
}

func (x *Controller) DeleteFile(ctx context.Context, in *service.ControllerDeleteFileIn) (*service.ControllerDeleteFileOut, error) {
	file, err := x.storage.GetFile(ctx, &repository.StorageGetFileIn{
		FileId: in.Id,
	})
	if err != nil {
		return nil, err
	}

	for _, shard := range file.Shards {
		getNode, err := x.infra.GetNode(ctx, &repository.InfraGetNodeIn{
			Id: shard.NodeId,
		})
		if err != nil {
			return nil, err
		}

		cli, err := x.getNodeClient(ctx, getNode.Node)
		if err != nil {
			return nil, err
		}

		filename := fmt.Sprintf("%s.%d", in.Id, shard.Index)

		if err = cli.DeleteFile(ctx, filename); err != nil {
			return nil, err
		}
	}

	if _, err = x.storage.DeleteFile(ctx, &repository.StorageDeleteFileIn{
		Id: file.Id,
	}); err != nil {
		return nil, err
	}

	return &service.ControllerDeleteFileOut{}, nil
}
