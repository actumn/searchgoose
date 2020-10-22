package discovery

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
	"sync"
)

type PreVoteCollector struct {
	state     map[state.Node]*PreVoteResponse //tuple
	preVotes  map[state.Node]*PreVoteResponse //map
	wirteLock sync.RWMutex

	electionStarted bool

	transportService  *transport.Service
	startElection     func()
	updateMaxTermSeen func(term int64)
}

func NewPreVoteCollector(transportService *transport.Service, startElection func(), updateMaxTermSeen func(term int64)) *PreVoteCollector {

	p := &PreVoteCollector{
		state:             make(map[state.Node]*PreVoteResponse),
		preVotes:          make(map[state.Node]*PreVoteResponse),
		transportService:  transportService,
		startElection:     startElection,
		updateMaxTermSeen: updateMaxTermSeen,
	}

	return p
}

func (p *PreVoteCollector) Start() {
	localNode := *(p.transportService.LocalNode)
	preVoteReqData := PreVoteRequest{
		SourceNode: localNode,
		Term:       p.getPreVoteResponse().CurrentTerm,
	}

	_, broadcastNodes := p.transportService.GetConnectedPeers()
	broadcastNodes = append(broadcastNodes, localNode)

	logrus.Infof("PreVoteCollector{SourceNode=%v} requesting pre-votes from %s\n", localNode, broadcastNodes)

	for _, node := range broadcastNodes {
		request := preVoteReqData.ToBytes()
		remoteNode := node
		logrus.Infof("PreVoteRequest%v to DestNode=%v\n", preVoteReqData, remoteNode)
		go p.transportService.SendRequest(node, transport.PREVOTE_REQ, request, func(res []byte) {
			data := PreVoteResponseFromBytes(res)
			logrus.Infof("PreVoteResponse%v from DestNode=%v\n", data, remoteNode)
			p.handlePreVoteResponse(data, remoteNode)
		})
	}
}

func (p *PreVoteCollector) handlePreVoteRequest(channel transport.ReplyChannel, req []byte) {

	preVoteReqData := PreVoteRequestFromBytes(req)
	logrus.Infof("PreVoteRequest=%v from DestNode={%s}\n", preVoteReqData, channel.GetDestAddress())

	p.updateMaxTermSeen(preVoteReqData.Term)

	leader := p.getLeader()
	preVoteResData := p.state[leader]

	// if the current node has not received the information that a node has been elected as the leader
	if leader != (state.Node{}) {
		logrus.Infof("Election already finished, won leader=%v", leader)
		preVoteResData.Err = "Election already finished"
	}

	response := preVoteResData.ToBytes()
	logrus.Infof("PreVoteResponse%v to DestNode={%s}", preVoteResData, channel.GetDestAddress())

	channel.SendMessage(transport.PREVOTE_RES, response)
}

func (p *PreVoteCollector) handlePreVoteResponse(response *PreVoteResponse, sender state.Node) {
	p.updateMaxTermSeen(response.CurrentTerm)
	p.SetVote(sender, response)

	voteCollection := state.NewVoteCollection()
	localNode := p.transportService.GetLocalNode()
	// localPreVoteResponse := p.PreVoteCollector.stateResponse

	for node, response := range p.preVotes {
		join := state.NewJoin(node, localNode, response.CurrentTerm)
		voteCollection.AddJoinVote(*join)
	}

	nodeIds, _ := p.transportService.GetConnectedPeers()
	nodeIds = append(nodeIds, localNode.Id)

	if voteCollection.IsQuorum(nodeIds) == false {
		logrus.Infof("No quorum yet")
		return
	}

	if p.electionStarted == true {
		logrus.Infof("Election already started")
		return
	}

	if p.getLeader() != (state.Node{}) {
		logrus.Infof("Already elected leader=%v", p.getLeader())
		return
	}

	p.electionStarted = true
	logrus.Infof("%v add %v from PrevoteResponse=%v\n, starting election\n", p.transportService.LocalNode, response, sender)

	//
	p.startElection()
}

func (p *PreVoteCollector) update(preVoteResponse *PreVoteResponse, leader state.Node) {
	if leader == (state.Node{}) {
		p.state[p.getLeader()] = preVoteResponse
	} else {
		delete(p.state, p.getLeader())
		p.state[leader] = preVoteResponse
	}
	logrus.Infof("Updating with preVoteResponse=%v, leader=%v\n", preVoteResponse, leader)

}

func (p *PreVoteCollector) SetVote(key state.Node, value *PreVoteResponse) {
	p.wirteLock.Lock()
	p.preVotes[key] = value
	p.wirteLock.Unlock()
}

func (p *PreVoteCollector) getLeader() state.Node {
	var leader state.Node
	for key, _ := range p.state {
		leader = key
	}
	return leader
}
func (p *PreVoteCollector) getPreVoteResponse() *PreVoteResponse {
	var response *PreVoteResponse
	for _, value := range p.state {
		response = value
	}
	return response
}

type PreVoteRequest struct {
	SourceNode state.Node
	Term       int64
}

func NewPreVoteRequest(sourceNode state.Node, term int64) *PreVoteRequest {
	return &PreVoteRequest{
		SourceNode: sourceNode,
		Term:       term,
	}
}

func (p *PreVoteRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(p); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func PreVoteRequestFromBytes(b []byte) *PreVoteRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PreVoteRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
}

type PreVoteResponse struct {
	CurrentTerm      int64
	lastAcceptedTerm int64
	// lastAcceptedVersion
	Err string
}

func NewPreVoteResponse(currentTerm int64) *PreVoteResponse {
	return &PreVoteResponse{
		CurrentTerm: currentTerm,
	}
}

func (p *PreVoteResponse) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(p); err != nil {
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func PreVoteResponseFromBytes(b []byte) *PreVoteResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PreVoteResponse
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}
