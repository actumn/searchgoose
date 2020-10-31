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
	preVotes map[state.Node]*PreVoteResponse

	leader          state.Node
	response        PreVoteResponse
	electionStarted bool

	transportService  *transport.Service
	startElection     func()
	updateMaxTermSeen func(term int64)

	Lock sync.RWMutex
}

func NewPreVoteCollector(transportService *transport.Service, startElection func(), updateMaxTermSeen func(term int64)) *PreVoteCollector {

	p := &PreVoteCollector{
		preVotes:          make(map[state.Node]*PreVoteResponse),
		transportService:  transportService,
		startElection:     startElection,
		updateMaxTermSeen: updateMaxTermSeen,
		Lock:              sync.RWMutex{},
	}

	return p
}

func (p *PreVoteCollector) Start() {
	localNode := p.transportService.GetLocalNode()

	request := PreVoteRequest{
		SourceNode: localNode,
		Term:       p.response.CurrentTerm,
	}

	_, broadcastNodes := p.transportService.GetConnectedPeers()
	broadcastNodes = append(broadcastNodes, localNode)

	logrus.Infof("PreVoteCollector: SourceNode=%v requesting pre-votes from %s\n", localNode, broadcastNodes)

	for _, node := range broadcastNodes {
		remoteNode := node
		logrus.Infof("PreVoteRequest=%v to DestNode=%v\n", request, remoteNode)
		// 이게 고루틴이어야 할까?
		p.transportService.SendRequest(node, transport.PREVOTE_REQ, request.ToBytes(), func(res []byte) {
			logrus.Info(56, res)
			data := PreVoteResponseFromBytes(res)
			logrus.Infof("PreVoteResponse%v from DestNode=%v\n", data, remoteNode)
			p.handlePreVoteResponse(data, remoteNode)
		})
	}
}

func (p *PreVoteCollector) handlePreVoteRequest(channel transport.ReplyChannel, req []byte) {

	request := PreVoteRequestFromBytes(req)
	logrus.Infof("PreVoteRequest=%v from DestNode={%s}\n", request, channel.GetDestAddress())

	p.updateMaxTermSeen(request.Term)

	response := p.response

	if p.leader != (state.Node{}) {
		logrus.Infof("Election already finished, won leader=%v", p.leader)
		response.Err = "Election already finished"
	}

	logrus.Infof("PreVoteResponse%v to DestNode={%s}", response, channel.GetDestAddress())

	channel.SendMessage(transport.PREVOTE_RES, response.ToBytes())
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

	if p.leader != (state.Node{}) {
		logrus.Infof("Already elected leader=%v", p.leader)
		return
	}

	p.electionStarted = true
	logrus.Infof("%v add %v from PrevoteResponse=%v\n, starting election\n", p.transportService.LocalNode, response, sender)

	//
	p.startElection()
}

func (p *PreVoteCollector) update(preVoteResponse *PreVoteResponse, leader state.Node) {

	p.leader = leader
	/*
		mutex := sync.RWMutex{}
		if leader == (state.Node{}) {
			mutex.Lock()
			p.state[p.getLeader()] = preVoteResponse
			mutex.Unlock()
		} else {
			mutex.Lock()
			delete(p.state, p.getLeader())
			p.state[leader] = preVoteResponse
			mutex.Unlock()
		}
	*/
	// logrus.Infof("Updating with preVoteResponse=%v, leader=%v\n", preVoteResponse, leader)

}

func (p *PreVoteCollector) SetVote(key state.Node, value *PreVoteResponse) {
	p.Lock.Lock()
	p.preVotes[key] = value
	p.Lock.Unlock()
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
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func PreVoteRequestFromBytes(b []byte) *PreVoteRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PreVoteRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}

type PreVoteResponse struct {
	CurrentTerm int64
	Err         string
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
