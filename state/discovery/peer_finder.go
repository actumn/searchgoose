package discovery

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"sync"
)

// public static final String REQUEST_PEERS_ACTION_NAME = "internal:discovery/request_peers";

type CoordinatorPeerFinder struct {
	Coordinator      *Coordinator
	transportService *transport.Service

	//mode        *Mode
	currentTerm       int64
	LastAcceptedNodes *state.Nodes

	//TODO: 얘가 과연 *state.Node여야 할까?
	PeersByAddress       map[string]*state.Node
	active               bool
	leader               *state.Node
	peersRequestInFlight bool
}

func NewCoordinatorPeerFinder(coordinator *Coordinator) *CoordinatorPeerFinder {
	f := &CoordinatorPeerFinder{
		Coordinator:          coordinator,
		transportService:     coordinator.TransportService,
		PeersByAddress:       make(map[string]*state.Node),
		peersRequestInFlight: false,
	}
	f.transportService.RegisterRequestHandler(transport.PEERFIND_REQ, f.handlePeersRequest)
	return f
}

func (f *CoordinatorPeerFinder) activate(lastAcceptedNodes *state.Nodes) {
	f.LastAcceptedNodes = lastAcceptedNodes
	f.active = true
	f.handleWakeUp()
}

func (f *CoordinatorPeerFinder) handleWakeUp() {
	// peer.handleWakeUp()

	providedAddr := f.getSeedHosts()

	var inner chan bool
	inner = make(chan bool, len(providedAddr))
	done := func() {
		inner <- true
	}

	wg := sync.WaitGroup{}
	wg.Add(len(providedAddr))
	for _, address := range providedAddr {
		if len(address) <= 0 {
			logrus.Warn("handleWakeUp: Invalid address")
			continue
		}
		logrus.Infof("handleWakeUp: Attempting connection to %s\n", address)
		go func(address string) {
			defer wg.Done()
			f.startProbe(address, done)
		}(address)
	}
	wg.Wait()
	// f.requestPeers(*remoteNode, f.onFoundPeersUpdated)
	//for i := 0 ; i < len(providedAddr); i++ {
	//	<- inner
	//}
	//
	//end()
}

func (f *CoordinatorPeerFinder) startProbe(address string, inner func()) {
	if _, ok := f.PeersByAddress[address]; !ok {
		f.createConnection(address, inner)
	}
}

func (f *CoordinatorPeerFinder) createConnection(address string, done func()) {
	leaderMutex := sync.Mutex{}
	f.transportService.ConnectToRemoteNode(address, func(remoteNode *state.Node) {
		if remoteNode == nil {
			// 만약 두 개의 seedhost가 있고, 하나는 커낵션이 안 되어 있고,
			// 나머지 하나가 master였다면?

			leaderMutex.Lock()
			f.Coordinator.becomeLeader("createConnection")
			leaderMutex.Unlock()
			//done()
			return
		}
		f.PeersByAddress[address] = remoteNode
		f.requestPeers(*remoteNode, f.Coordinator.startPreVote, done)
	})
}

func (f *CoordinatorPeerFinder) requestPeers(destNode state.Node, next func(), done func()) {
	nowNode := f.getLocalNode()
	foundPeers := f.getFoundPeers()

	request := PeersRequest{
		SourceNode: nowNode,
		KnownPeers: foundPeers,
	}

	f.peersRequestInFlight = true

	logrus.Infof("RequestPeers: Peer=%v requesting peers %s\n", nowNode, foundPeers)

	f.transportService.SendRequest(destNode, transport.PEERFIND_REQ, request.ToBytes(), func(res []byte) {
		data := PeersResponseFromBytes(res)
		master := data.MasterNode
		peers := data.KnownPeers
		term := data.Term

		logrus.Infof("RequestPeers: Peer=%v received PeersResponse=%v\n", nowNode, data)

		//if f.active == false {
		//	return
		//}

		f.peersRequestInFlight = false

		if master != (state.Node{}) {
			if master == destNode {
				f.onActiveMasterFound(destNode, term, done)
			} else {
				f.startProbe(master.HostAddress, func() {})
				f.Coordinator.becomeFollower("requestPeers", master)
			}
		}

		for _, peer := range peers {
			f.transportService.ConnectToRemoteNode(peer.HostAddress, func(remoteNode *state.Node) {
				f.requestPeers(*remoteNode, func() {}, func() {})
			})
		}

		// start election
		// next()

		done()
	})
}

func (f *CoordinatorPeerFinder) handlePeersRequest(channel transport.ReplyChannel, req []byte) {
	request := PeersRequestFromBytes(req)
	peers := request.KnownPeers
	logrus.Infof("Receive Peer Finding REQ from %s; %s\n", channel.GetDestAddress(), peers)

	response := PeersResponse{
		MasterNode: state.Node{},
		KnownPeers: []state.Node{},
	}

	// if f.active == true

	if f.leader != nil {
		response.MasterNode = *(f.leader)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(peers))
	for _, peer := range peers {
		go func(address string) {
			defer wg.Done()
			f.startProbe(address, func() {})
		}(peer.HostAddress)
	}
	wg.Wait()

	knownPeers := f.getFoundPeers()
	logrus.Infof("Send Peer Finding RES to %s; %s\n", channel.GetDestAddress(), knownPeers)
	response.KnownPeers = knownPeers
	response.Term = f.currentTerm

	channel.SendMessage(transport.PEERFIND_ACK, response.ToBytes())
}

func (f *CoordinatorPeerFinder) onActiveMasterFound(leader state.Node, term int64, done func()) {
	// ensureTermAtLeast(masterNode, term);
	// joinHelper.sendJoinRequest(masterNode, getCurrentTerm(), joinWithDestination(lastJoin, masterNode, term));

	f.Coordinator.ensureTermAtLeast(leader, term)
	f.Coordinator.JoinHelper.SendJoinRequest(leader, f.Coordinator.getCurrentTerm(), nil, done)

}

func (f *CoordinatorPeerFinder) getFoundPeers() []state.Node {
	//ids := make([]string, 0, len(f.PeersByAddress))
	values := make([]state.Node, 0, len(f.PeersByAddress))

	for _, v := range f.PeersByAddress {
		//ids = append(ids, v.Id)
		values = append(values, *v)
	}
	return values
	//return ids, values
}

func (f *CoordinatorPeerFinder) getLocalNode() state.Node {
	return *(f.transportService.LocalNode)
}

func (f *CoordinatorPeerFinder) getSeedHosts() []string {
	return f.transportService.GetSeedHosts()
}

type PeersRequest struct {
	SourceNode state.Node
	KnownPeers []state.Node
}

func (r *PeersRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func PeersRequestFromBytes(b []byte) *PeersRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}

type PeersResponse struct {
	MasterNode state.Node
	KnownPeers []state.Node
	Term       int64
}

func (r *PeersResponse) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func PeersResponseFromBytes(b []byte) *PeersResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersResponse
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}
