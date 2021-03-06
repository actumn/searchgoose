package discovery

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport"
	"github.com/sirupsen/logrus"
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
	h.transportService.SendRequest(destination, transport.START_JOIN_REQ, request, func(res []byte) {
		logrus.Infof("StartJoinRequest : successful response=%v from %v\n", startJoinRequest, destination)
	})
}

func (h *JoinHelper) handleStartJoinRequest(channel transport.ReplyChannel, req []byte) {
	startJoinReqData := StartJoinRequestFromBytes(req)
	destination := startJoinReqData.SourceNode

	join := h.joinLeaderInTerm(startJoinReqData)

	h.SendJoinRequest(destination, h.currentTermSupplier(), join)

	channel.SendMessage(transport.START_JOIN_ACK, []byte("Send START_JOIN_ACK"))
}

func (h *JoinHelper) SendJoinRequest(destination state.Node, term int64, join *state.Join) {

	var newJoin state.Join
	if join == nil {
		newJoin = state.Join{
			Term: 1,
		}
	} else {
		newJoin = *join
	}

	joinRequest := JoinRequest{
		SourceNode:  h.transportService.GetLocalNode(),
		MinimumTerm: term,
		Join:        newJoin,
	}

	logrus.Infof("SendJoinRequest: Attempting to join=%v with joinRequest=%v\n", destination, joinRequest)

	request := joinRequest.ToBytes()

	remoteAddress := destination.HostAddress

	h.transportService.ConnectToRemoteNode(remoteAddress, func(node *state.Node) {
		h.transportService.SendRequest(*node, transport.JOIN_REQ, request, func(res []byte) {
			logrus.Infof("Successfully joined %v with %v\n", destination, joinRequest)
		})
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
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func StartJoinRequestFromBytes(b []byte) *StartJoinRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data StartJoinRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
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
		logrus.Fatalln(err)
	}
	return buffer.Bytes()
}

func (r *JoinRequest) GetTerm() int64 {
	return common.GetMaxInt(r.MinimumTerm, r.Join.Term)
}

func JoinRequestFromBytes(b []byte) *JoinRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data JoinRequest
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatalln(err)
	}
	return &data
}
