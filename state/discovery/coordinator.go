package discovery

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
)

type Mode int

const (
	INIT Mode = iota
	CANDIDATE
	LEADER
	FOLLOWER
	PREVOTING
)

// Coordinator
type Coordinator struct {
	TransportService *transport.Service
	PeerFinder       *CoordinatorPeerFinder
	PersistedState   state.PersistedState

	CoordinationState       state.CoordinationState
	ApplierState            *state.ClusterState
	ClusterApplierService   *cluster.ApplierService
	MasterService           *cluster.MasterService
	ClusterBootstrapService *ClusterBootstrapService

	PreVoteCollector *PreVoteCollector

	JoinHelper *JoinHelper
	lastJoin   *state.Join

	maxTermSeen int64

	mode Mode
}

func NewCoordinator(transportService *transport.Service, clusterApplierService *cluster.ApplierService, masterService *cluster.MasterService, persistedState state.PersistedState) *Coordinator {
	c := &Coordinator{
		TransportService:      transportService,
		ClusterApplierService: clusterApplierService,
		MasterService:         masterService,
		PersistedState:        persistedState,
		maxTermSeen:           1,
	}

	c.TransportService.RegisterRequestHandler("publish_state", c.handlePublish)

	c.PreVoteCollector = NewPreVoteCollector(transportService, c.startElection, c.updateMaxTermSeen)
	c.TransportService.RegisterRequestHandler(transport.PREVOTE_REQ, c.PreVoteCollector.handlePreVoteRequest)

	c.JoinHelper = NewJoinHelper(transportService, c.joinLeaderInTerm, c.getCurrentTerm, c.handleJoinRequest)
	c.TransportService.RegisterRequestHandler(transport.START_JOIN, c.JoinHelper.handleStartJoinRequest)
	c.TransportService.RegisterRequestHandler(transport.JOIN_REQ, c.handleJoinRequest)

	return c
}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		PersistedState: c.PersistedState,
		Term:           1,
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
	c.PeerFinder = NewCoordinatorPeerFinder(c)
	c.PeerFinder.currentTerm = c.getCurrentTerm()

	c.PreVoteCollector.state[state.Node{}] = NewPreVoteResponse(c.getCurrentTerm())

	c.ClusterApplierService.ClusterState = c.ApplierState
	c.MasterService.ClusterState = c.ApplierState
}

func (c *Coordinator) StartInitialJoin() {
	c.becomeCandidate("start_initial")
	// clusterBootstrapService.scheduleUnconfiguredBootstrap();
}

func (c *Coordinator) becomeCandidate(method string) {
	if c.mode != CANDIDATE {
		c.mode = CANDIDATE
		c.PeerFinder.activate(c.CoordinationState.PersistedState.GetLastAcceptedState().Nodes)
	}
}

func (c *Coordinator) becomeLeader(method string) {
	logrus.Printf("%v: coordinator becoming LEADER in term {%d}\n", method, c.getCurrentTerm())
	c.mode = LEADER
	// preVoteCollector.update(getPreVoteResponse(), getLocalNode())
}

func (c *Coordinator) becomeFollower(method string, leaderNode state.Node) {

}

func (c *Coordinator) updateMaxTermSeen(term int64) {

	c.maxTermSeen = common.GetMaxInt(c.maxTermSeen, term)
	currentTerm := c.getCurrentTerm()

	if c.mode == LEADER && c.maxTermSeen > currentTerm {

	}
}

func (c *Coordinator) startElection() {
	localNode := *(c.TransportService.LocalNode)

	if c.mode == PREVOTING {
		startJoinRequest := StartJoinRequest{
			SourceNode: localNode,
			Term:       common.GetMaxInt(c.maxTermSeen, 1) + 1,
		}

		logrus.Info("Start election with %v\n", startJoinRequest)

		// discoveredNodes := c.PeerFinder.getFoundPeers()
		_, discoveredNodes := c.TransportService.GetConnectedPeers()
		for _, node := range discoveredNodes {
			go c.JoinHelper.SendStartJoinRequest(startJoinRequest, node)
		}
	}
}

func (c *Coordinator) joinLeaderInTerm(request *StartJoinRequest) *state.Join {
	localNode := *(c.TransportService.LocalNode)

	logrus.Printf("joinLeaderInTerm: for [{%v}] with term {%d}\n", request.SourceNode, request.Term)
	if request.Term <= c.getCurrentTerm() {
		logrus.Printf("handleStartJoin: ignoring as term provided is not greater than current term \n")
	}

	logrus.Printf("handleStartJoin: leaving term [%d] due to %v", c.getCurrentTerm(), request)

	c.CoordinationState.Term = request.Term
	c.CoordinationState.JoinVotes = state.NewVoteCollection()
	c.CoordinationState.PublishVotes = state.NewVoteCollection()

	join := state.NewJoin(localNode, request.SourceNode, c.getCurrentTerm())
	c.lastJoin = join

	c.PeerFinder.currentTerm = c.getCurrentTerm()

	if c.mode != PREVOTING {
		c.becomeCandidate("join_leader_in_term")
	} else {
		// followersChecker.updateFastResponseState(getCurrentTerm(), mode);
		c.PreVoteCollector.update(NewPreVoteResponse(c.getCurrentTerm()), state.Node{})
	}

	return join
}

func (c *Coordinator) handleJoinRequest(channel transport.ReplyChannel, req []byte) {
	// transportService.connectToNode -> 대체 왜?
	joinReqData := JoinRequestFromBytes(req)
	logrus.Printf("handleJoinRequest: as {%d}, handling %v\n", c.mode, joinReqData)
	c.updateMaxTermSeen(joinReqData.MinimumTerm)
	c.handleJoin(joinReqData.Join)
	//  joinAccumulator.handleJoinRequest(joinRequest.getSourceNode(), joinCallback);
	if c.CoordinationState.ElectionWon == true {
		c.becomeLeader("handleJoinRequest")
	}
}

func (c *Coordinator) handleJoin(join state.Join) {
	// ensureTermAtLeast(getLocalNode(), join.getTerm()).ifPresent(this::handleJoin);
	// localNode := c.TransportService.GetLocalNode()

	if c.CoordinationState.ElectionWon == false {
		if c.getCurrentTerm() != join.Term {
			logrus.Println("handleJoin: ignored join due to term mismatch")
		}
		c.CoordinationState.JoinVotes.AddJoinVote(join)

		nodeIds, _ := c.TransportService.GetConnectedPeers()
		// nodeIds = append(nodeIds, localNode.Id)
		c.CoordinationState.ElectionWon = c.CoordinationState.IsElectionQuorum(nodeIds)

		if c.CoordinationState.ElectionWon {
			logrus.Printf("handleJoin: election won in term [{%d}] with {%v}\n", c.getCurrentTerm(), c.CoordinationState.JoinVotes)
		}
	}
}

func (c *Coordinator) getCurrentTerm() int64 {
	return c.CoordinationState.Term
}

func (c *Coordinator) ensureTermAtLeast(discoveryNode state.Node, targetTerm int64) {
	if c.getCurrentTerm() < targetTerm {

	}
}

func (c *Coordinator) Publish(event state.ClusterChangedEvent) {
	//if c.mode != LEADER {
	//	return
	//}

	newState := event.State
	nodes := newState.Nodes

	for _, node := range nodes.Nodes {
		logrus.Info("publish new state to node[" + node.Id + "]")
		c.TransportService.SendRequest(*node, "publish_state", newState.ToBytes(), func(response []byte) {

		})
	}
	logrus.Info("publish ended successfully")
}

func (c *Coordinator) handlePublish(channel transport.ReplyChannel, req []byte) {
	// handle publish
	acceptedState := state.ClusterStateFromBytes(req, c.TransportService.LocalNode)
	//localState := c.CoordinationState.PersistedState.GetLastAcceptedState()
	logrus.Info("accept new state ")
	c.CoordinationState.PersistedState.SetLastAcceptedState(acceptedState)

	// handle commit
	c.ApplierState = acceptedState
	c.MasterService.ClusterState = acceptedState
	c.ClusterApplierService.OnNewState(acceptedState)

	channel.SendMessage("publish_start", []byte{})
}

// PeerFinder
type CoordinatorPeerFinder struct {
	Coordinator *Coordinator

	mode        *Mode
	currentTerm int64

	transportService  *transport.Service
	LastAcceptedNodes *state.Nodes
	PeersByAddress    map[string]*state.Node
	active            bool

	leader *state.Node
}

func NewCoordinatorPeerFinder(coordinator *Coordinator) *CoordinatorPeerFinder {
	f := &CoordinatorPeerFinder{
		Coordinator:      coordinator,
		transportService: coordinator.TransportService,
		PeersByAddress:   make(map[string]*state.Node),
	}

	return f
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *state.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
	// f.onFoundPeersUpdated()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {
	providedAddr := f.transportService.Transport.GetSeedHosts()
	for _, address := range providedAddr {
		logrus.Info("Attempting connection to %s\n", address)
		go f.startProbe(address)
	}
}

func (f *CoordinatorPeerFinder) startProbe(address string /*wg *sync.WaitGroup*/) {
	if _, ok := f.PeersByAddress[address]; !ok {
		f.createConnection(address)
	}
}

func (f *CoordinatorPeerFinder) createConnection(address string) {
	f.transportService.ConnectToRemoteNode(address, func(remoteNode *state.Node) {
		f.PeersByAddress[address] = remoteNode

		_, foundPeers := f.transportService.GetConnectedPeers()
		f.transportService.RequestPeers(*remoteNode, foundPeers, f.onFoundPeersUpdated)
	})
}

func (f *CoordinatorPeerFinder) onFoundPeersUpdated() {
	if f.Coordinator.mode == CANDIDATE {
		f.Coordinator.mode = PREVOTING
		f.startElectionScheduler()
	} else {
		logrus.Println("PreVoting already stared")
	}
}

func (f *CoordinatorPeerFinder) startElectionScheduler() {
	if f.Coordinator.mode == PREVOTING {
		f.Coordinator.PreVoteCollector.Start()
	}
}

func (f *CoordinatorPeerFinder) deactivate(leader *state.Node) {
	logrus.Printf("Deactivating and setting leader to %v\n", leader)
	f.active = false
	// peersRemoved = handleWakeUp();
	f.leader = leader

	/*
		if (peersRemoved) {
			onFoundPeersUpdated();
		}
	*/

}

func (f *CoordinatorPeerFinder) getFoundPeers() ([]string, []state.Node) {
	ids := make([]string, 0, len(f.PeersByAddress))
	values := make([]state.Node, 0, len(f.PeersByAddress))

	for _, v := range f.PeersByAddress {
		ids = append(ids, v.Id)
		values = append(values, *v)
	}

	return ids, values
}
