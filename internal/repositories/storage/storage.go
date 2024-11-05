package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/fydmer/fileserver/internal/domain/repository"
	"github.com/fydmer/fileserver/internal/errors/pgerr"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) (*Repository, error) {
	return &Repository{db: db}, nil
}

func (r *Repository) CreateFile(ctx context.Context, in *repository.StorageCreateFileIn) (*repository.StorageCreateFileOut, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return nil, pgerr.Parse(err)
	}

	var fileId string

	fileQuery := `insert into files (location) values ($1) returning id`
	if err = tx.QueryRowContext(ctx, fileQuery, in.Location).Scan(&fileId); err != nil {
		return nil, pgerr.Parse(errors.Join(err, tx.Rollback()))
	}

	for _, shard := range in.Shards {
		shardQuery := `insert into shards (file_id, node_id, index, size, status) values ($1, $2, $3, $4, $5)`
		if _, err = tx.ExecContext(ctx, shardQuery, fileId, shard.NodeId, shard.Index, shard.Size, repository.StorageShardStatusNew); err != nil {
			return nil, pgerr.Parse(errors.Join(err, tx.Rollback()))
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, pgerr.Parse(err)
	}

	return &repository.StorageCreateFileOut{
		Id: fileId,
	}, nil
}

func (r *Repository) SetShardStatus(ctx context.Context, in *repository.StorageSetShardStatusIn) (*repository.StorageSetShardStatusOut, error) {
	query := `update shards set status = $4 where file_id = $1 and node_id = $2 and index = $3`

	_, err := r.db.ExecContext(ctx, query, in.FileId, in.NodeId, in.Index, in.Status)
	if err != nil {
		return nil, pgerr.Parse(err)
	}

	return &repository.StorageSetShardStatusOut{}, nil
}

func (r *Repository) GetFile(ctx context.Context, in *repository.StorageGetFileIn) (*repository.StorageGetFileOut, error) {
	query := `select f.location, s.node_id, s.index, s.size, s.created_at, s.status 
    from shards s inner join files f on f.id = s.file_id where file_id = $1`

	rows, err := r.db.QueryContext(ctx, query, in.FileId)
	if err != nil {
		return nil, pgerr.Parse(err)
	}

	defer rows.Close()

	var location string
	var shards []*repository.StorageShard
	for rows.Next() {
		shard := &repository.StorageShard{}
		if err = rows.Scan(&location, &shard.NodeId, &shard.Index, &shard.Size, &shard.CreatedAt, &shard.Status); err != nil {
			return nil, pgerr.Parse(err)
		}
		shards = append(shards, shard)
	}

	return &repository.StorageGetFileOut{
		Id:       in.FileId,
		Location: location,
		Shards:   shards,
	}, nil
}

func (r *Repository) GetFileByLocation(ctx context.Context, in *repository.StorageGetFileByLocationIn) (*repository.StorageGetFileByLocationOut, error) {
	query := `select id from files where lower(location) = lower($1)`

	var id string
	if err := r.db.QueryRowContext(ctx, query, in.Location).Scan(&id); err != nil {
		return nil, pgerr.Parse(err)
	}

	return r.GetFile(ctx, &repository.StorageGetFileIn{FileId: id})
}

func (r *Repository) DeleteFile(ctx context.Context, in *repository.StorageDeleteFileIn) (*repository.StorageDeleteFileOut, error) {
	query := `delete from files where id = $1`

	if _, err := r.db.ExecContext(ctx, query, in.Id); err != nil {
		return nil, pgerr.Parse(err)
	}

	return &repository.StorageDeleteFileOut{}, nil
}
