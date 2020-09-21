package discovery

import (
	"fmt"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/transport"
	"log"
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
		PeerFinder:            NewCoordinatorPeerFinder(transportService),
		ClusterApplierService: clusterApplierService,
		MasterService:         masterService,
		PersistedState:        persistedState,
	}

	// c.TransportService.RegisterRequestHandler("publish_state", c.HandlePublish)

	return c
}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		PersistedState: c.PersistedState,
	}

	//c.ApplierState = c.PersistedState.GetLastAcceptedState()
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
	c.becomeCandidate("start_initial")
}

func (c *Coordinator) becomeCandidate(method string) {
	if c.mode != CANDIDATE {
		c.mode = CANDIDATE
		c.PeerFinder.activate(c.CoordinationState.PersistedState.GetLastAcceptedState().Nodes)
	}
}

func (c *Coordinator) Publish(event state.ClusterChangedEvent) {
	c.mode = LEADER
	if c.mode != LEADER {
		return
	}

	newState := event.State
	nodes := newState.Nodes

	for _, node := range nodes.Nodes {
		// c.TransportService.SendRequest(*node, "publish_state", newState.ToBytes())
		fmt.Println(node)
	}
}

func (c *Coordinator) HandlePublish(req []byte) []byte {
	// handle publish
	acceptedState := state.ClusterStateFromBytes(req, c.TransportService.LocalNode)
	//localState := c.CoordinationState.PersistedState.GetLastAcceptedState()

	c.CoordinationState.PersistedState.SetLastAcceptedState(acceptedState)

	// handle commit
	c.ApplierState = acceptedState
	c.MasterService.ClusterState = acceptedState
	c.ClusterApplierService.OnNewState(acceptedState)

	return []byte{}
}

type CoordinatorPeerFinder struct {
	transportService  *transport.Service
	LastAcceptedNodes *state.Nodes
	PeersByAddress    map[string]*state.Node
	active            bool
}

func NewCoordinatorPeerFinder(transportService *transport.Service) *CoordinatorPeerFinder {
	f := &CoordinatorPeerFinder{
		transportService: transportService,
		PeersByAddress:   make(map[string]*state.Node),
	}
	return f
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *state.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {
	providedAddr := f.transportService.Transport.SeedHosts
	for _, address := range providedAddr {
		f.startProbe(address)
	}
}

func (f *CoordinatorPeerFinder) startProbe(address string) {
	if _, ok := f.PeersByAddress[address]; !ok {
		peer := f.createConnectingPeer(address)
		f.PeersByAddress[address] = peer
	}
}

func (f *CoordinatorPeerFinder) createConnectingPeer(address string) *state.Node {
	// 여기서 Peer 만드셈
	log.Printf("Attempting connection to %s\n", address)
	return f.establishConnection(address)
}

func (f *CoordinatorPeerFinder) establishConnection(address string) *state.Node {
	// Handshaking
	node := &state.Node{}
	f.transportService.ConnectToRemoteMasterNode(address, func(remoteNode state.Node) {
		// Peer Finding
		node = &remoteNode
		knownPeers := f.getConnectedPeers()
		log.Printf("[Known Peers] %s\n", knownPeers)

		remotePeers := f.transportService.RequestPeers(remoteNode, knownPeers)

		for _, peer := range remotePeers {
			f.startProbe(peer.HostAddress)
		}

	})
	return node
}

func (f *CoordinatorPeerFinder) getConnectedPeers() []state.Node {
	values := make([]state.Node, 0, len(f.transportService.ConnectionManager))
	for _, v := range f.PeersByAddress {
		values = append(values, *v)
	}
	return values
}
