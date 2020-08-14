package discovery

import (
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/transport"
)

type Mode int

const (
	INIT Mode = iota
	CANDIDATE
	LEADER
	FOLLOWER
)

type Coordinator struct {
	TransportService transport.Service
	PeerFinder       CoordinatorPeerFinder
	PersistedState   state.PersistedState

	CoordinationState     state.CoordinationState
	ApplierState          state.ClusterState
	ClusterApplierService cluster.ApplierService

	mode Mode
}

func newCoordinator() {

}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		PersistedState: c.PersistedState,
	}

	c.ApplierState = state.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &state.Nodes{
			Nodes: map[string]*state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.LocalNode,
			},
			LocalNodeId: c.TransportService.LocalNode.Id,
		},
	}
	c.ClusterApplierService.ClusterState = c.ApplierState
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
	LastAcceptedNodes *state.Nodes
	active            bool
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *state.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {

}

func (f *CoordinatorPeerFinder) onFoundPeersUpdated() {

}
