package transport

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"log"
	"net"
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
	SendRequest(req []byte, callback func(byte []byte))
}

type Transport interface {
	OpenConnection(address string, callback func(conn Connection))
	Start(address string)
	Register(action string, handler RequestHandler)
	GetLocalAddress() string
	GetSeedHosts() []string
}

// structures
type RequestHandler func(conn net.Conn, req []byte) []byte

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
	service.RegisterRequestHandler(HANDSHAKE_REQ, service.handleHandShake)
	service.RegisterRequestHandler(PEERFIND_REQ, service.HandlePeersRequest)
	return service
}

func (s *Service) Start() {
	address := s.Transport.GetLocalAddress()
	s.Transport.Start(address)

	time.Sleep(time.Duration(15) * time.Second)
}

func (s *Service) SendRequestConn(conn Connection, action string, req []byte, callback func(response []byte)) {
	//// local node
	// handler := s.Transport.RequestHandlers[action]
	// handler(req)

	conn.SendRequest(req, callback)
}

func (s *Service) SendRequest(node state.Node, action string, req []byte, callback func(response []byte)) {
	conn := s.ConnectionManager[node]
	s.SendRequestConn(conn, action, req, callback)
}

func (s *Service) RegisterRequestHandler(action string, handler RequestHandler) {
	s.Transport.Register(action, handler)
}

func (s *Service) ConnectToRemoteMasterNode(address string, callback func(node state.Node)) {

	nowNode := s.LocalNode
	connectedNode := state.Node{}

	s.Transport.OpenConnection(address, func(conn Connection) {
		handShakeData := DataFormat{
			Source:  nowNode.HostAddress,
			Dest:    address,
			Action:  HANDSHAKE_REQ,
			Content: nowNode.ToBytes(),
		}
		log.Printf("Send handshake REQ to %s\n", address)
		request := handShakeData.ToBytes()
		s.SendRequestConn(conn, HANDSHAKE_REQ, request, func(response []byte) {
			node := state.NodeFromBytes(response)
			connectedNode = *node
			s.ConnectionManager[connectedNode] = conn
		})
	})

	log.Printf("Connected with  %s\n", address)
	time.Sleep(time.Duration(10) * time.Second)

	callback(connectedNode)

}

func (s *Service) handleHandShake(conn net.Conn, req []byte) []byte {
	reqNode := state.NodeFromBytes(req)
	log.Printf("Receive handshake REQ from %s\n", reqNode.HostAddress)

	nowNode := s.LocalNode
	handShakeData := DataFormat{
		Source:  nowNode.HostAddress,
		Action:  HANDSHAKE_ACK,
		Content: nowNode.ToBytes(),
	}

	return handShakeData.ToBytes()
}

func (s *Service) RequestPeers(node state.Node, knownPeers []state.Node) []state.Node {
	// knownNodes => peersByAddress에서 가져오렴

	content := PeersRequest{
		SourceNode: *s.LocalNode,
		KnownPeers: knownPeers,
	}

	nowNode := s.LocalNode
	log.Printf("Send Peer Finding REQ to %s\n", node.HostAddress)
	peerFindData := DataFormat{
		Source:  nowNode.HostAddress,
		Dest:    node.HostAddress,
		Action:  PEERFIND_REQ,
		Content: content.ToBytes(),
	}

	// TODO :: 나중에 request handler interface로 뽑아내기
	request := peerFindData.ToBytes()
	var peers []state.Node
	s.SendRequest(node, PEERFIND_REQ, request, func(response []byte) {
		peers = PeersResponseFromBytes(response).KnownPeers
		log.Printf("%s received %s", s.LocalNode.HostAddress, peers)
	})

	return peers
}

func (s *Service) HandlePeersRequest(conn net.Conn, req []byte) []byte {
	request := PeersRequestFromBytes(req)
	log.Printf("Receive Peer Finding REQ from %s\n", request.SourceNode.Id)

	knownPeers := make([]state.Node, 0, len(s.ConnectionManager))
	for peer := range s.ConnectionManager {
		knownPeers = append(knownPeers, peer)
	}

	nowNode := s.LocalNode
	response := PeersResponse{
		KnownPeers: knownPeers,
	}

	handShakeData := DataFormat{
		Source:  nowNode.HostAddress,
		Action:  PEERFIND_ACK,
		Content: response.ToBytes(),
	}

	return handShakeData.ToBytes()
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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func DataFormatFromBytes(b []byte) *DataFormat {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data DataFormat
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func PeersRequestFromBytes(b []byte) *PeersRequest {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersRequest
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
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
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

func PeersResponseFromBytes(b []byte) *PeersResponse {
	buffer := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buffer)
	var data PeersResponse
	if err := decoder.Decode(&data); err != nil {
		log.Fatalln(err)
	}
	return &data
}

/*
type LocalConnection struct {
}
func (c *LocalConnection) GetNode() {
}
func (c *LocalConnection) SendRequest(req []byte) {
	//handler.sdfsadfsadfsadf
}
*/
