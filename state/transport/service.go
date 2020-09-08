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
	requestHandlers   map[string]RequestHandler
	ConnectionManager map[string]Connection
	transport         *tcp.Transport
}

func NewService(id string, transport *tcp.Transport) *Service {
	return &Service{
		LocalNode:       state.CreateLocalNode(id),
		requestHandlers: map[string]RequestHandler{},
		transport:       transport,
	}
}

func (s *Service) Start() {
	address := s.transport.LocalAddress
	s.transport.Start(address)

	time.Sleep(time.Duration(15) * time.Second)

	log.Printf("Start handshaking\n")

	seedHosts := s.transport.SeedHosts
	connections := make(chan *net.Conn)
	for _, seedHost := range seedHosts {
		// Open connection
		s.transport.OpenConnection(seedHost, connections)
	}

	time.Sleep(time.Duration(20) * time.Second)

}

func (s *Service) NodeConnected(node *state.Node) bool {
	// TODO :: 만약 profile로 seedHost가 주어졌다면, connectionManager에 저장하기
	// return isLocalNode(node) || connectionManager.nodeConnected(node);
	_, found := s.ConnectionManager[node.Id]
	return node == s.LocalNode || found
}

func (s *Service) SendRequest(node state.Node, action string, req []byte) {
	// local node
	handler := s.requestHandlers[action]
	handler(req)
	// TODO :: send request to remote node
}

func (s *Service) RegisterRequestHandler(action string, handler RequestHandler) {
	s.requestHandlers[action] = handler
}

func (s *Service) openConnection() {

}

type RequestHandler func(req []byte)

/*
type Transport interface {
}
*/

type Connection interface {
	getNode()
	sendRequest()
}
