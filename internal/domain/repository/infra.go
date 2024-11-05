package repository

import "context"

type InfraNode struct {
	Id   string
	Addr string
}

type InfraCreateNodeIn struct {
	Addr string
}

type InfraCreateNodeOut struct {
	Node *InfraNode
}

type InfraGetNodeIn struct {
	Id string
}

type InfraGetNodeOut struct {
	Node *InfraNode
}

type InfraListNodesIn struct{}

type InfraListNodesOut struct {
	Nodes []*InfraNode
}

type InfraGetFreerNodesIn struct {
	Count int
}

type InfraGetFreerNodesOut struct {
	Nodes []*InfraNode
}

type Infra interface {
	CreateNode(ctx context.Context, in *InfraCreateNodeIn) (*InfraCreateNodeOut, error)
	GetNode(ctx context.Context, in *InfraGetNodeIn) (*InfraGetNodeOut, error)
	ListNodes(ctx context.Context, in *InfraListNodesIn) (*InfraListNodesOut, error)
	GetFreerNodes(ctx context.Context, in *InfraGetFreerNodesIn) (*InfraGetFreerNodesOut, error)
}
