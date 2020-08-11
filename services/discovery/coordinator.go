package discovery

import (
	"github.com/actumn/searchgoose/services/metadata"
	"github.com/actumn/searchgoose/services/persist"
	"github.com/actumn/searchgoose/services/transport"
)

type Coordinator struct {
	transportService transport.Service
	PeerFinder       PeerFinder
	persistedState   persist.PersistedState

	CoordinationState CoordinationState
	ApplierState      metadata.ClusterState
}

func (c *Coordinator) Start() {
	c.CoordinationState = CoordinationState{
		LocalNode:      c.transportService.LocalNode,
		PersistedState: c.persistedState,
	}

	c.ApplierState = metadata.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: Nodes{
			Nodes: map[string]*Node{
				c.transportService.LocalNode.Id: c.transportService.LocalNode,
			},
			LocalNodeId: c.transportService.LocalNode.Id,
		},
	}
}

func (c *Coordinator) StartInitialJoin() {

}

func (c *Coordinator) becomeCandidate() {

}

type PeerFinder interface {
}

type JoinHelper interface {
}
