package discovery

import (
	"github.com/actumn/searchgoose/services"
	"github.com/actumn/searchgoose/services/transport"
)

type Mode int

const (
	INIT Mode = iota
	CANDIDATE
	LEADER
	FOLLOWER
)

type Coordinator struct {
	transportService transport.Service
	PeerFinder       CoordinatorPeerFinder
	PersistedState   services.PersistedState

	CoordinationState services.CoordinationState
	ApplierState      services.ClusterState

	mode Mode
}

func newCoordinator() {

}

func (c *Coordinator) Start() {
	c.CoordinationState = services.CoordinationState{
		LocalNode:      c.transportService.LocalNode,
		PersistedState: c.PersistedState,
	}

	c.ApplierState = services.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &services.Nodes{
			Nodes: map[string]*services.Node{
				c.transportService.LocalNode.Id: c.transportService.LocalNode,
			},
			LocalNodeId: c.transportService.LocalNode.Id,
		},
	}
}

func (c *Coordinator) StartInitialJoin() {
	c.becomeCandidate("startInital")
}

func (c *Coordinator) becomeCandidate(method string) {
	if c.mode != CANDIDATE {
		c.mode = CANDIDATE
		c.PeerFinder.activate(c.CoordinationState.PersistedState.GetLastAcceptedState().Nodes)
	}
}

type CoordinatorPeerFinder struct {
	LastAcceptedNodes *services.Nodes
	active            bool
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *services.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {

}

func (f *CoordinatorPeerFinder) onFoundPeersUpdated() {

}
