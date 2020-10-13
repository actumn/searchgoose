package discovery

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
	"log"
)

type PreVoteCollector struct {
	state           map[state.Node]*PreVoteResponse //tuple
	preVotes        map[state.Node]*PreVoteResponse //map
	electionStarted bool

	transportService  *transport.Service
	startElection     func()
	updateMaxTermSeen func(term int64)
}

func NewPreVoteCollector(transportService *transport.Service, startElection func(), updateMaxTermSeen func(term int64)) *PreVoteCollector {

	p := &PreVoteCollector{
		state:    make(map[state.Node]*PreVoteResponse),
		preVotes: make(map[state.Node]*PreVoteResponse),

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
	// broadcastNodes = append(broadcastNodes, localNode)

	log.Printf("PreVoteCollector{SourceNode=%v} requesting pre-votes from %s\n", localNode, broadcastNodes)

	for _, node := range broadcastNodes {
		request := preVoteReqData.ToBytes()
		remoteNode := node
		log.Printf("PreVoteRequest%v to DestNode=%v\n", preVoteReqData, remoteNode)
		go p.transportService.SendRequest(node, transport.PREVOTE_REQ, request, func(res []byte) {
			data := PreVoteResponseFromBytes(res)
			log.Printf("PreVoteResponse%v from DestNode=%v\n", data, remoteNode)
			p.handlePreVoteResponse(data, remoteNode)
		})
	}
}

func (p *PreVoteCollector) handlePreVoteRequest(channel transport.ReplyChannel, req []byte) {

	preVoteReqData := PreVoteRequestFromBytes(req)
	log.Printf("PreVoteRequest%v from DestNode={%s}\n", preVoteReqData, channel.GetDestAddress())

	p.updateMaxTermSeen(preVoteReqData.Term)

	leader := p.getLeader()
	preVoteResData := p.state[leader]

	// if the current node has not received the information that a node has been elected as the leader
	if leader != (state.Node{}) {
		//return error
	}

	response := preVoteResData.ToBytes()
	log.Printf("PreVoteResponse%v to DestNode={%s}\n", preVoteResData, channel.GetDestAddress())

	channel.SendMessage(transport.PREVOTE_RES, response)
}

func (p *PreVoteCollector) handlePreVoteResponse(response *PreVoteResponse, sender state.Node) {
	p.updateMaxTermSeen(response.CurrentTerm)
	p.preVotes[sender] = response

	voteCollection := state.NewVoteCollection()
	localNode := *(p.transportService.LocalNode)
	// localPreVoteResponse := p.PreVoteCollector.stateResponse

	for node, response := range p.preVotes {
		join := state.NewJoin(node, localNode, response.CurrentTerm)
		voteCollection.AddJoinVote(*join)
	}

	nodeIds, _ := p.transportService.GetConnectedPeers()

	if voteCollection.IsQuorum(nodeIds) == false {
		fmt.Println("No quorum yet")
		return
	}

	if p.electionStarted == true {
		fmt.Println("Election already started")
		return
	}

	p.electionStarted = true
	log.Printf("%v add %v from %v\n, starting election\n", p.transportService.LocalNode, response, sender)

	p.startElection()
}

func (p *PreVoteCollector) update(preVoteResponse *PreVoteResponse, leader state.Node) {
	if leader == (state.Node{}) {
		p.state[p.getLeader()] = preVoteResponse
	} else {
		delete(p.state, p.getLeader())
		p.state[leader] = preVoteResponse
	}
	log.Printf("Updating with preVoteResponse=%v, leader=%v\n", preVoteResponse, leader)

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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func PreVoteRequestFromBytes(b []byte) *PreVoteRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PreVoteRequest
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}

type PreVoteResponse struct {
	CurrentTerm      int64
	lastAcceptedTerm int64
	// lastAcceptedVersion
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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func PreVoteResponseFromBytes(b []byte) *PreVoteResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PreVoteResponse
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}
