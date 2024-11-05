package infra

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

func (r *Repository) CreateNode(ctx context.Context, in *repository.InfraCreateNodeIn) (*repository.InfraCreateNodeOut, error) {
	query := `insert into nodes (addr) values ($1) returning id`
	node := &repository.InfraNode{
		Addr: in.Addr,
	}

	if err := r.db.QueryRowContext(ctx, query, in.Addr).Scan(&node.Id); err != nil {
		return nil, pgerr.Parse(err)
	}

	return &repository.InfraCreateNodeOut{
		Node: node,
	}, nil
}

func (r *Repository) GetNode(ctx context.Context, in *repository.InfraGetNodeIn) (*repository.InfraGetNodeOut, error) {
	query := `select id, addr from nodes where id=$1`
	node := &repository.InfraNode{}
	if err := r.db.QueryRowContext(ctx, query, in.Id).Scan(&node.Id, &node.Addr); err != nil {
		return nil, pgerr.Parse(err)
	}
	return &repository.InfraGetNodeOut{
		Node: node,
	}, nil
}

func (r *Repository) ListNodes(ctx context.Context, _ *repository.InfraListNodesIn) (*repository.InfraListNodesOut, error) {
	query := `select id, addr from nodes`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &repository.InfraListNodesOut{}, nil
		}
		return nil, pgerr.Parse(err)
	}
	defer rows.Close()

	var nodes []*repository.InfraNode
	for rows.Next() {
		node := &repository.InfraNode{}
		if err := rows.Scan(&node.Id, &node.Addr); err != nil {
			return nil, pgerr.Parse(err)
		}
		nodes = append(nodes, node)
	}
	return &repository.InfraListNodesOut{
		Nodes: nodes,
	}, nil
}

func (r *Repository) GetFreerNodes(ctx context.Context, in *repository.InfraGetFreerNodesIn) (*repository.InfraGetFreerNodesOut, error) {
	query := `select n.id, n.addr from nodes n left join shards s on n.id = s.node_id 
    where status not in ($2) or status is null group by (n.id, n.addr) order by coalesce(sum(size), 0) limit $1`
	rows, err := r.db.QueryContext(ctx, query, in.Count, repository.StorageShardStatusError)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &repository.InfraGetFreerNodesOut{}, nil
		}
		return nil, pgerr.Parse(err)
	}
	defer rows.Close()

	var nodes []*repository.InfraNode
	for rows.Next() {
		node := &repository.InfraNode{}
		if err := rows.Scan(&node.Id, &node.Addr); err != nil {
			return nil, pgerr.Parse(err)
		}
		nodes = append(nodes, node)
	}

	return &repository.InfraGetFreerNodesOut{
		Nodes: nodes,
	}, nil
}
