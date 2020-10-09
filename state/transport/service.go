package transport

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	HANDSHAKE_REQ  = "handshake_req"
	HANDSHAKE_ACK  = "handshake_ack"
	HANDSHAKE_FAIL = "handshake_fail"
	PEERFIND_REQ   = "peerfind_req"
	PEERFIND_ACK   = "peerfind_ack"
	PEERFIND_FAIL  = "peerfind_fail"
)

// interfaces
type Connection interface {
	SendRequest(action string, req []byte, callback func(byte []byte))
}

type ReplyChannel interface {
	SendMessage(b []byte) (n int, err error)
}

type Transport interface {
	OpenConnection(address string, callback func(conn Connection))
	Start(address string)
	Register(action string, handler RequestHandler)
	GetLocalAddress() string
	GetSeedHosts() []string
	GetHandler(action string) RequestHandler
}

// structures
type RequestHandler func(channel ReplyChannel, req []byte) []byte

type Service struct {
	LocalNode *state.Node
	// ConnectionManager map[string]Connection
	ConnectionManager map[state.Node]Connection
	Transport         Transport
}

func NewService(id string, transport Transport) *Service {
	address := transport.GetLocalAddress()
	service := &Service{
		LocalNode:         state.CreateLocalNode(id, address),
		ConnectionManager: make(map[state.Node]Connection),
		Transport:         transport,
	}
	service.RegisterRequestHandler(HANDSHAKE_REQ, service.handleHandshake)
	service.RegisterRequestHandler(PEERFIND_REQ, service.HandlePeersRequest)
	return service
}

func (s *Service) Start() {
	address := s.Transport.GetLocalAddress()
	s.Transport.Start(address)

	// time.Sleep(time.Duration(15) * time.Second)
}

func (s *Service) SendRequestConn(conn Connection, action string, req []byte, callback func(response []byte)) {
	conn.SendRequest(action, req, callback)
}

func (s *Service) SendRequest(node state.Node, action string, req []byte, callback func(response []byte)) {
	var conn Connection
	if node.Id == s.LocalNode.Id {
		conn = &LocalConnection{
			service: s,
		}
	} else {
		conn = s.ConnectionManager[node]
	}

	s.SendRequestConn(conn, action, req, callback)
}

func (s *Service) RegisterRequestHandler(action string, handler RequestHandler) {
	s.Transport.Register(action, handler)
}

func (s *Service) ConnectToRemoteMasterNode(address string, callback func(node state.Node)) {
	nowNode := s.LocalNode
	connectedNode := state.Node{}

	s.Transport.OpenConnection(address, func(conn Connection) {
		handshakeData := DataFormat{
			Source:  nowNode.HostAddress,
			Dest:    address,
			Action:  HANDSHAKE_REQ,
			Content: nowNode.ToBytes(),
		}
		logrus.Info("Send handshake REQ to %s", address)
		request := handshakeData.ToBytes()
		s.SendRequestConn(conn, HANDSHAKE_REQ, request, func(response []byte) {
			node := state.NodeFromBytes(response)
			connectedNode = *node
			s.ConnectionManager[connectedNode] = conn
		})
	})

	logrus.Info("Connected with  %s", address)
	time.Sleep(time.Duration(10) * time.Second)

	callback(connectedNode)

}

func (s *Service) handleHandshake(channel ReplyChannel, req []byte) []byte {
	reqNode := state.NodeFromBytes(req)
	logrus.Info("Receive handshake REQ from %s", reqNode.HostAddress)

	nowNode := s.LocalNode
	handshakeData := DataFormat{
		Source:  nowNode.HostAddress,
		Action:  HANDSHAKE_ACK,
		Content: nowNode.ToBytes(),
	}

	return handshakeData.ToBytes()
}

func (s *Service) RequestPeers(node state.Node, knownPeers []state.Node) []state.Node {
	// knownNodes => peersByAddress에서 가져오렴

	content := PeersRequest{
		SourceNode: *s.LocalNode,
		KnownPeers: knownPeers,
	}

	nowNode := s.LocalNode
	logrus.Info("Send Peer Finding REQ to %s", node.HostAddress)
	peerFindData := DataFormat{
		Source:  nowNode.HostAddress,
		Dest:    node.HostAddress,
		Action:  PEERFIND_REQ,
		Content: content.ToBytes(),
	}

	logrus.Info("[%s] %s", PEERFIND_REQ, content)

	// TODO :: 나중에 request handler interface로 뽑아내기
	request := peerFindData.ToBytes()
	var peers []state.Node
	s.SendRequest(node, PEERFIND_REQ, request, func(response []byte) {
		peers = PeersResponseFromBytes(response).KnownPeers
		logrus.Info("%s received %s", s.LocalNode.HostAddress, peers)
	})

	return peers
}

func (s *Service) HandlePeersRequest(channel ReplyChannel, req []byte) []byte {
	request := PeersRequestFromBytes(req)
	logrus.Info("Receive Peer Finding REQ from %s", request.SourceNode.Id)

	knownPeers := make([]state.Node, 0, len(s.ConnectionManager))
	for peer := range s.ConnectionManager {
		knownPeers = append(knownPeers, peer)
	}

	nowNode := s.LocalNode
	response := PeersResponse{
		KnownPeers: knownPeers,
	}

	handshakeData := DataFormat{
		Source:  nowNode.HostAddress,
		Action:  PEERFIND_ACK,
		Content: response.ToBytes(),
	}

	return handshakeData.ToBytes()
}

// Data types
type DataFormat struct {
	Source  string
	Dest    string
	Action  string
	Content []byte
}

func (d *DataFormat) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(d); err != nil {
		logrus.Fatal(err)
	}
	return buffer.Bytes()
}

func DataFormatFromBytes(b []byte) *DataFormat {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data DataFormat
	if err := decoder.Decode(&data); err != nil {
		logrus.Fatal(err)
	}
	return &data
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

type LocalConnection struct {
	service *Service
}

func (c *LocalConnection) SendRequest(action string, req []byte, callback func(byte []byte)) {
	handler := c.service.Transport.GetHandler(action)
	replyChannel := &DirectReplyChannel{
		callback: callback,
	}

	data := handler(replyChannel, req)
	replyChannel.SendMessage(data)
}

type DirectReplyChannel struct {
	callback func(byte []byte)
}

func (c *DirectReplyChannel) SendMessage(b []byte) (int, error) {
	c.callback(b)
	return 0, nil
}
