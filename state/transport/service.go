package transport

import (
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

func (s *Service) SendRequest(node *state.Node, action string, req []byte, callback func(response []byte)) {
	conn := s.ConnectionManager[node.Id]
	s.SendRequestConn(conn, action, req, callback)
}

func (s *Service) RegisterRequestHandler(action string, handler tcp.RequestHandler) {
	s.Transport.Register(action, handler)
}

func (s *Service) ConnectToRemoteMasterNode(address string) *state.Node {

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
	return &connectedNode
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
