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
	PeerFinder       *CoordinatorPeerFinder
	PersistedState   state.PersistedState

	CoordinationState     state.CoordinationState
	ApplierState          *state.ClusterState
	ClusterApplierService *cluster.ApplierService
	MasterService         *cluster.MasterService

	mode Mode
}

func NewCoordinator(transportService *transport.Service, clusterApplierService *cluster.ApplierService, masterService *cluster.MasterService, persistedState state.PersistedState) *Coordinator {
	c := &Coordinator{
		TransportService:      transportService,
		ClusterApplierService: clusterApplierService,
		MasterService:         masterService,
		PeerFinder:            NewCoordinatorPeerFinder(transportService),
	}
	c.TransportService.RegisterRequestHandler("publish_state", c.HandlePublish)

	return c
}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		PersistedState: c.PersistedState,
	}

	c.ApplierState = c.PersistedState.GetLastAcceptedState()
	//c.ApplierState = &state.ClusterState{
	//	Name: "searchgoose-testClusters",
	//	Nodes: &state.Nodes{
	//		Nodes: map[string]*state.Node{
	//			c.TransportService.LocalNode.Id: c.TransportService.LocalNode,
	//		},
	//		LocalNodeId: c.TransportService.LocalNode.Id,
	//	},
	//}
	c.ClusterApplierService.ClusterState = c.ApplierState
	c.MasterService.ClusterState = c.ApplierState
}

func (c *Coordinator) StartInitialJoin() {
	c.becomeCandidate("startInitial")
}

func (c *Coordinator) becomeCandidate(method string) {
	if c.mode != CANDIDATE {
		c.mode = CANDIDATE
		c.PeerFinder.activate(c.CoordinationState.PersistedState.GetLastAcceptedState().Nodes)
	}
}

func (c *Coordinator) Publish(event state.ClusterChangedEvent) {
	if c.mode != LEADER {
		return
	}

	newState := event.State
	nodes := newState.Nodes

	for _, node := range nodes.Nodes {
		c.TransportService.SendRequest(*node, "publish_state", newState.ToBytes())
	}
}

func (c *Coordinator) HandlePublish(req []byte) {
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
	TransportService  *transport.Service
	LastAcceptedNodes *state.Nodes
	active            bool
	peersByAddress    map[string]*Peer
}

func NewCoordinatorPeerFinder(transportService *transport.Service) *CoordinatorPeerFinder {
	return &CoordinatorPeerFinder{
		TransportService: transportService,
	}
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *state.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {
	// "localhost:8080" 기준
	seedHosts := []string{"localhost:8081", "localhost:8082"}

	for _, host := range seedHosts {
		f.startProbe(host)
	}

}

func (f *CoordinatorPeerFinder) startProbe(address string) {
	for key, _ := range f.peersByAddress {
		if key == address {
			return
		}
	}
	peer := f.createConnectingPeer(address)
	f.peersByAddress[address] = peer
}

func (f *CoordinatorPeerFinder) createConnectingPeer(address string) *Peer {
	// f.TransportService.

	/*
		peer := NewPeer(address)
		peer.establishConnection()
		return peer
	*/
}

func (f *CoordinatorPeerFinder) onFoundPeersUpdated() {

}

type Peer struct {
	address       string
	discoveryNode state.Node
}

func NewPeer(address string) *Peer {
	return &Peer{
		address: address,
	}
}

func (p *Peer) establishConnection() {

}
