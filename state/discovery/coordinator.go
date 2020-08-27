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
	TransportService *transport.Service
	PeerFinder       CoordinatorPeerFinder
	PersistedState   state.PersistedState

	CoordinationState     state.CoordinationState
	ApplierState          *state.ClusterState
	ClusterApplierService *cluster.ApplierService
	MasterService         *cluster.MasterService

	mode Mode
}

func NewCoordinator() {
	c := Coordinator{}
	c.TransportService.RegisterRequestHandler("publish_state", c.handlePublish)
}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		PersistedState: c.PersistedState,
	}

	c.ApplierState = &state.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &state.Nodes{
			Nodes: map[string]*state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.LocalNode,
			},
			LocalNodeId: c.TransportService.LocalNode.Id,
		},
	}
	c.ClusterApplierService.ClusterState = c.ApplierState
	c.MasterService.ClusterState = c.ApplierState
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

func (c *Coordinator) publish(event state.ClusterChangedEvent) {
	if c.mode != LEADER {
		return
	}

	newState := event.State
	nodes := newState.Nodes

	for _, node := range nodes.Nodes {
		c.TransportService.SendRequest(*node, "publish_state", newState.ToBytes())
	}
}

func (c *Coordinator) handlePublish(req []byte) {
	// handle publish
	acceptedState := state.ClusterStateFromBytes(req, c.TransportService.LocalNode)
	//localState := c.CoordinationState.PersistedState.GetLastAcceptedState()

	c.CoordinationState.PersistedState.SetLastAcceptedState(acceptedState)

	// handle commit
	c.ApplierState = acceptedState
	c.MasterService.ClusterState = acceptedState
	c.ClusterApplierService.OnNewState(acceptedState)
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
