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

	JoinHelper      *JoinHelper
	lastJoin        *state.Join
	lastKnownLeader *state.Node

	maxTermSeen int64

	mode Mode

	Done func()
}

func NewCoordinator(transportService *transport.Service, clusterApplierService *cluster.ApplierService, masterService *cluster.MasterService, persistedState state.PersistedState) *Coordinator {
	c := &Coordinator{
		TransportService:      transportService,
		ClusterApplierService: clusterApplierService,
		MasterService:         masterService,
		PersistedState:        persistedState,
		maxTermSeen:           1,
	}

	c.TransportService.RegisterRequestHandler(transport.PUBLISH_REQ, c.handlePublish)

	c.PreVoteCollector = NewPreVoteCollector(transportService, c.startElection, c.updateMaxTermSeen)
	c.TransportService.RegisterRequestHandler(transport.PREVOTE_REQ, c.PreVoteCollector.handlePreVoteRequest)

	c.JoinHelper = NewJoinHelper(transportService, c.joinLeaderInTerm, c.getCurrentTerm, c.handleJoinRequest)
	c.TransportService.RegisterRequestHandler(transport.START_JOIN_REQ, c.JoinHelper.handleStartJoinRequest)
	c.TransportService.RegisterRequestHandler(transport.JOIN_REQ, c.handleJoinRequest)

	return c
}

func (c *Coordinator) Start() {
	c.CoordinationState = state.CoordinationState{
		LocalNode:      c.TransportService.LocalNode,
		JoinVotes:      state.NewVoteCollection(),
		PersistedState: c.PersistedState,
		Term:           1,
	}

	c.ApplierState = c.PersistedState.GetLastAcceptedState()
	c.ApplierState = &state.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &state.Nodes{
			Nodes: map[string]state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.GetLocalNode(),
			},
			LocalNodeId: c.TransportService.LocalNode.Id,
			DataNodes: map[string]state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.GetLocalNode(),
			},
			MasterNodes: map[string]state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.GetLocalNode(),
			},
			MasterNodeId: c.TransportService.LocalNode.Id,
		},
		Metadata: state.Metadata{
			Indices:       map[string]state.IndexMetadata{},
			IndicesLookup: map[string]state.IndexAbstractionAlias{},
		},
	}
	c.PeerFinder = NewCoordinatorPeerFinder(c)
	c.PeerFinder.currentTerm = c.getCurrentTerm()

	c.PreVoteCollector.state[state.Node{}] = NewPreVoteResponse(c.getCurrentTerm())

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

func (c *Coordinator) becomeLeader(method string) {

	if method == "createConnection" && c.mode == LEADER {
		logrus.Infof("becomeLeader: elected as LEADER before")
		return
	}

	logrus.Infof("%v: Coordinator becoming LEADER in term {%d}\n", method, c.getCurrentTerm())
	localNode := c.TransportService.GetLocalNode()
	c.mode = LEADER
	c.PeerFinder.deactivate(localNode)
	c.PreVoteCollector.update(NewPreVoteResponse(c.getCurrentTerm()+1), localNode)

	// make new cluster state
	newClusterState := &state.ClusterState{
		Name: "searchgoose-testClusters",
		Nodes: &state.Nodes{
			LocalNodeId:  c.TransportService.LocalNode.Id,
			MasterNodeId: c.TransportService.LocalNode.Id,
			Nodes: map[string]state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.GetLocalNode(),
			},
			DataNodes: map[string]state.Node{},
			MasterNodes: map[string]state.Node{
				c.TransportService.LocalNode.Id: c.TransportService.GetLocalNode(),
			},
		},
		Metadata: state.Metadata{
			Indices:       map[string]state.IndexMetadata{},
			IndicesLookup: map[string]state.IndexAbstractionAlias{},
		},
	}

	_, nodes := c.TransportService.GetConnectedPeers()
	nodes = append(nodes, c.TransportService.GetLocalNode())
	for _, node := range nodes {
		newClusterState.Nodes.Nodes[node.Id] = node
		newClusterState.Nodes.DataNodes[node.Id] = node
		newClusterState.Nodes.MasterNodes[node.Id] = node
	}

	c.Publish(state.ClusterChangedEvent{
		State:     *newClusterState,
		PrevState: *(c.ApplierState),
	})

}

func (c *Coordinator) becomeFollower(method string, leaderNode state.Node) {

}

func (c *Coordinator) updateMaxTermSeen(term int64) {
	logrus.Infof("updateMaxTermSeen: maxTermSeen={%d} term={%d}", c.maxTermSeen, term)
	c.maxTermSeen = common.GetMaxInt(c.maxTermSeen, term)
	currentTerm := c.getCurrentTerm()

	if c.mode == LEADER && c.maxTermSeen > currentTerm {
		// leader exit
	}
}

func (c *Coordinator) startElection() {
	localNode := *(c.TransportService.LocalNode)

	if c.mode == PREVOTING {
		startJoinRequest := StartJoinRequest{
			SourceNode: localNode,
			Term:       common.GetMaxInt(c.maxTermSeen, 1) + 1,
		}

		logrus.Infof("Start election with %v\n", startJoinRequest)

		// discoveredNodes := c.PeerFinder.getFoundPeers()
		_, discoveredNodes := c.TransportService.GetConnectedPeers()
		discoveredNodes = append(discoveredNodes, c.TransportService.GetLocalNode())
		for _, node := range discoveredNodes {
			go c.JoinHelper.SendStartJoinRequest(startJoinRequest, node)
		}
	}
}

func (c *Coordinator) joinLeaderInTerm(request *StartJoinRequest) *state.Join {
	localNode := *(c.TransportService.LocalNode)

	logrus.Infof("joinLeaderInTerm: for %v with term={%d}\n", request.SourceNode, request.Term)
	if request.Term <= c.getCurrentTerm() {
		logrus.Infof("handleStartJoin: ignoring as term provided is not greater than current term \n")
	}

	logrus.Infof("handleStartJoin: leaving term [%d] due to %v", c.getCurrentTerm(), request)

	c.CoordinationState.Term = request.Term
	c.CoordinationState.JoinVotes = state.NewVoteCollection()
	//c.CoordinationState.PublishVotes = state.NewVoteCollection()

	join := state.NewJoin(localNode, request.SourceNode, c.getCurrentTerm())
	c.lastJoin = join

	c.PeerFinder.currentTerm = c.getCurrentTerm()

	if c.mode != PREVOTING {
		//c.becomeCandidate("joinLeaderInTerm")
	} else {
		// followersChecker.updateFastResponseState(getCurrentTerm(), mode);
		c.PreVoteCollector.update(NewPreVoteResponse(c.getCurrentTerm()), c.ApplierState.Nodes.MasterNode())
		//c.PreVoteCollector.update(NewPreVoteResponse(c.getCurrentTerm()), state.Node{})
	}

	return join
}

func (c *Coordinator) handleJoinRequest(channel transport.ReplyChannel, req []byte) {
	c.TransportService.ConnectToRemoteNode(channel.GetDestAddress(), func(remoteNode *state.Node) {
		joinReqData := JoinRequestFromBytes(req)
		logrus.Infof("handleJoinRequest: as {%d}, handling %v\n", c.mode, joinReqData)
		c.updateMaxTermSeen(joinReqData.GetTerm())
		c.handleJoin(joinReqData.Join)
		//  joinAccumulator.handleJoinRequest(joinRequest.getSourceNode(), joinCallback);
		if c.CoordinationState.ElectionWon == true {
			c.becomeLeader("handleJoinRequest")
		}
	})
}

func (c *Coordinator) handleJoin(join state.Join) {
	localNode := c.TransportService.GetLocalNode()
	localJoin := c.ensureTermAtLeast(localNode, join.Term)
	if localJoin != nil {
		c.handleJoin(*localJoin)
	}

	if c.CoordinationState.ElectionWon == false {
		if c.getCurrentTerm() != join.Term {
			logrus.Infof("handleJoin: ignored join due to term mismatch current={%d} term={%d}", c.getCurrentTerm(), join.Term)
			return
		}

		c.CoordinationState.ElectionWon = true
		logrus.Infof("handleJoin: election won in term={%d} with %v\n", c.getCurrentTerm(), c.CoordinationState.JoinVotes)

		//c.CoordinationState.JoinVotes.AddJoinVote(join)
		//
		//nodeIds, _ := c.TransportService.GetConnectedPeers()
		//nodeIds = append(nodeIds, localNode.Id)
		//c.CoordinationState.ElectionWon = c.CoordinationState.IsElectionQuorum(nodeIds)
		//
		//if c.CoordinationState.ElectionWon {
		//	logrus.Infof("handleJoin: election won in term={%d} with %v\n", c.getCurrentTerm(), c.CoordinationState.JoinVotes)
		//}
	}
}

func (c *Coordinator) getCurrentTerm() int64 {
	return c.CoordinationState.Term
}

func (c *Coordinator) ensureTermAtLeast(sourceNode state.Node, targetTerm int64) *state.Join {
	if c.getCurrentTerm() < targetTerm {
		request := &StartJoinRequest{
			SourceNode: sourceNode,
			Term:       targetTerm,
		}
		return c.joinLeaderInTerm(request)
	}
	return nil
}

func (c *Coordinator) Publish(event state.ClusterChangedEvent) {
	//if c.mode != LEADER {
	//	return
	//}

	newState := event.State
	nodes := newState.Nodes.Nodes
	for _, node := range nodes {
		logrus.Infof("Publish: leader=%v publish to DestNode=%v", c.TransportService.LocalNode, node)
		c.TransportService.SendRequest(node, transport.PUBLISH_REQ, newState.ToBytes(), func(response []byte) {

		})
	}
	logrus.Info("publish ended successfully")
}

func (c *Coordinator) handlePublish(channel transport.ReplyChannel, req []byte) {
	// handle publish
	acceptedState := state.ClusterStateFromBytes(req, c.TransportService.GetLocalNode())
	//localState := c.CoordinationState.PersistedState.GetLastAcceptedState()
	logrus.Info("accept new state from leader=%v", acceptedState.Nodes.MasterNode())
	c.CoordinationState.PersistedState.SetLastAcceptedState(acceptedState)

	// handle commit
	c.ApplierState = acceptedState
	c.MasterService.ClusterState = acceptedState
	c.ClusterApplierService.OnNewState(acceptedState)

	for _, node := range acceptedState.Nodes.Nodes {
		go c.TransportService.ConnectToRemoteNode(node.HostAddress, func(node *state.Node) {

		})
	}

	c.Done()

	channel.SendMessage(transport.PUBLISH_ACK, []byte{})
}

func (c *Coordinator) startPreVote() {
	if c.mode != CANDIDATE {
		c.mode = PREVOTING
		c.PreVoteCollector.Start()
	} else {
		logrus.Infof("PreVoting already stared")
		return
	}
}

func (f *CoordinatorPeerFinder) deactivate(leader state.Node) {
	logrus.Infof("Deactivating and setting leader to %v\n", leader)
	f.active = false
	// peersRemoved = f.handleWakeUp
	f.leader = &leader

	/*
		if (peersRemoved) {
			onFoundPeersUpdated();
		}
	*/

}
