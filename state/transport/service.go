package transport

import (
	"bytes"
	"encoding/gob"
	"github.com/actumn/searchgoose/state"
	"github.com/actumn/searchgoose/state/transport/tcp"
	"log"
	"net"
	"time"
)

type Service struct {
	LocalNode         *state.Node
	ConnectionManager map[string]Connection
	Transport         *tcp.Transport
}

func NewService(id string, transport *tcp.Transport) *Service {
	address := transport.LocalAddress
	service := &Service{
		LocalNode: state.CreateLocalNode(id, address),
		Transport: transport,
	}
	service.RegisterRequestHandler(tcp.HANDSHAKE_REQ, service.handleHandShake)
	service.RegisterRequestHandler(tcp.PEERFIND_REQ, service.HandlePeersRequest)
	return service
}

func (s *Service) Start() {
	address := s.Transport.LocalAddress
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
	conn := s.ConnectionManager[node.Id]
	s.SendRequestConn(conn, action, req, callback)
}

func (s *Service) RegisterRequestHandler(action string, handler tcp.RequestHandler) {
	s.Transport.Register(action, handler)
}

func (s *Service) ConnectToRemoteMasterNode(address string, callback func(node state.Node)) {

	nowNode := s.LocalNode
	connectedNode := state.Node{}

	s.Transport.OpenConnection(address, func(conn Connection) {
		handShakeData := tcp.DataFormat{
			Source:  nowNode.HostAddress,
			Dest:    address,
			Action:  tcp.HANDSHAKE_REQ,
			Content: nowNode.ToBytes(),
		}
		request := handShakeData.ToBytes()
		s.SendRequestConn(conn, tcp.HANDSHAKE_REQ, request, func(response []byte) {
			node := state.NodeFromBytes(response)
			connectedNode = *node
			s.ConnectionManager[node.Id] = conn
		})
	})
	callback(connectedNode)
}

func (s *Service) handleHandShake(conn net.Conn, req []byte) []byte {
	reqNode := state.NodeFromBytes(req)
	log.Printf("Receive handshake REQ from %s\n", reqNode.Id)

	nowNode := s.LocalNode
	handShakeData := tcp.DataFormat{
		Source:  nowNode.HostAddress,
		Action:  tcp.HANDSHAKE_ACK,
		Content: nowNode.ToBytes(),
	}

	return handShakeData.ToBytes()
}

func (s *Service) RequestPeers(node state.Node, knownPeers []state.Node) []state.Node {
	// knownNodes => peersByAddress에서 가져오렴

	content := PeersRequest{
		knownPeers: knownPeers,
	}

	nowNode := s.LocalNode
	peerFindData := tcp.DataFormat{
		Source:  nowNode.HostAddress,
		Dest:    node.HostAddress,
		Action:  tcp.PEERFIND_REQ,
		Content: content.ToBytes(),
	}

	// TODO :: 나중에 request handler interface로 뽑아내기
	request := peerFindData.ToBytes()
	var peers []state.Node
	s.SendRequest(node, tcp.PEERFIND_REQ, request, func(response []byte) {
		peers = PeersResponseFromBytes(response).knownPeers
		log.Printf("%s received %s", s.LocalNode.HostAddress, peers)
		// response.getMasterNode().map(DiscoveryNode::getAddress).ifPresent(PeerFinder.this::startProbe);
		// response.getKnownPeers().stream().map(DiscoveryNode::getAddress).forEach(PeerFinder.this::startProbe);
	})
	return peers
}

func (s *Service) HandlePeersRequest(conn net.Conn, req []byte) []byte {

}

type PeersRequest struct {
	knownPeers []state.Node
}

func (r *PeersRequest) ToBytes() []byte {
	var buffer bytes.Buffer
	enc := gob.NewEncoder(&buffer)
	if err := enc.Encode(r); err != nil {
		log.Fatalln(err)
	}
	return buffer.Bytes()
}

type PeersResponse struct {
	masterNode state.Node
	knownPeers []state.Node
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
func (s *Service) openConnection() {

}

/*
type Transport interface {
}
*/

type Connection interface {
	SendRequest(req []byte, callback func(byte []byte))
}

type LocalConnection struct {
}

func (c *LocalConnection) GetNode() {

}

func (c *LocalConnection) SendRequest(req []byte) {
	//handler.sdfsadfsadfsadf
}
