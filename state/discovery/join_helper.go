package discovery

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
	"log"
)

type JoinHelper struct {
	transportService *transport.Service

	// functions from coordinator
	joinLeaderInTerm    func(request *StartJoinRequest) *state.Join
	currentTermSupplier func() int64
	joinHandler         func(channel transport.ReplyChannel, req []byte)
}

func NewJoinHelper(
	transportService *transport.Service,
	joinLeaderInTerm func(request *StartJoinRequest) *state.Join,
	currentTermSupplier func() int64,
	joinHandler func(channel transport.ReplyChannel, req []byte)) *JoinHelper {
	return &JoinHelper{
		transportService:    transportService,
		joinLeaderInTerm:    joinLeaderInTerm,
		currentTermSupplier: currentTermSupplier,
		joinHandler:         joinHandler,
	}
}

func (h *JoinHelper) SendStartJoinRequest(startJoinRequest StartJoinRequest, destination state.Node) {
	request := startJoinRequest.ToBytes()
	h.transportService.SendRequest(destination, transport.START_JOIN, request, func(res []byte) {
		log.Printf("Successful response to %v from %v\n", startJoinRequest, destination)
	})
}

func (h *JoinHelper) handleStartJoinRequest(channel transport.ReplyChannel, req []byte) {
	startJoinReqData := StartJoinRequestFromBytes(req)
	destination := startJoinReqData.SourceNode

	join := h.joinLeaderInTerm(startJoinReqData)

	h.SendJoinRequest(destination, h.currentTermSupplier(), join)

	channel.SendMessage(transport.START_JOIN_ACK, []byte{})
}

func (h *JoinHelper) SendJoinRequest(destination state.Node, term int64, join *state.Join) {

	joinRequest := JoinRequest{
		SourceNode:  h.transportService.GetLocalNode(),
		MinimumTerm: term,
		Join:        *join,
	}

	log.Printf("Attempting to join %v with %v\n", destination, joinRequest)

	request := joinRequest.ToBytes()

	remoteAddress := destination.HostAddress
	if h.transportService.IsConnected(remoteAddress) == false {
		h.transportService.ConnectToRemoteNode(remoteAddress, func(node *state.Node) {})
	}

	h.transportService.SendRequest(destination, transport.JOIN_REQ, request, func(res []byte) {
		log.Printf("Successfully joined %v with %v\n", destination, joinRequest)
	})
}

type JoinAccumulator interface {
	handleJoinRequest(sender state.Node)
}

type InitialJoinAccumulator struct {
}

func (a *InitialJoinAccumulator) handleJoinRequest(sender state.Node) {

}

type CandidateJoinAccumulator struct {
}

//Data Format
type StartJoinRequest struct {
	SourceNode state.Node
	Term       int64
}

func (r *StartJoinRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func StartJoinRequestFromBytes(b []byte) *StartJoinRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data StartJoinRequest
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}

type JoinRequest struct {
	SourceNode  state.Node
	MinimumTerm int64
	Join        state.Join
}

func (r *JoinRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func JoinRequestFromBytes(b []byte) *JoinRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data JoinRequest
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}
