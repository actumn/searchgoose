package discovery

import (
	"github.com/actumn/searchgoose/services/metadata"
	"github.com/actumn/searchgoose/services/persist"
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
	persistedState   persist.PersistedState

	CoordinationState CoordinationState
	ApplierState      metadata.ClusterState

	mode Mode
}

func newCoordinator() {

}

func (c *Coordinator) Start() {
	c.CoordinationState = CoordinationState{
		LocalNode:      c.transportService.LocalNode,
		PersistedState: c.persistedState,
	}

	c.ApplierState = metadata.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &Nodes{
			Nodes: map[string]*Node{
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
	LastAcceptedNodes *Nodes
	active            bool
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {

}

func (f *CoordinatorPeerFinder) onFoundPeersUpdated() {

}
