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
	RequestHandlers   map[string]RequestHandler
	ConnectionManager map[string]Connection
	Transport         *tcp.Transport
}

func NewService(id string, transport *tcp.Transport) *Service {
	return &Service{
		LocalNode:       state.CreateLocalNode(id),
		RequestHandlers: map[string]RequestHandler{},
		Transport:       transport,
	}
}

func (s *Service) Start() {
	address := s.Transport.LocalAddress
	s.Transport.Start(address)

	time.Sleep(time.Duration(15) * time.Second)

	log.Printf("Start handshaking\n")

	seedHosts := s.Transport.SeedHosts
	connections := make(chan *net.Conn)
	for _, seedHost := range seedHosts {
		// Open connection
		s.Transport.OpenConnection(seedHost, connections)
	}

	time.Sleep(time.Duration(20) * time.Second)

}

func (s *Service) SendRequest(node state.Node, action string, req []byte) {
	// local node
	handler := s.RequestHandlers[action]
	handler(req)
	// TODO :: send request to remote node
}

func (s *Service) RegisterRequestHandler(action string, handler RequestHandler) {
	s.RequestHandlers[action] = handler
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
