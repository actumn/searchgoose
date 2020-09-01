package transport

import (
	"github.com/actumn/searchgoose/state"
)

type Service struct {
	LocalNode         *state.Node
	requestHandlers   map[string]RequestHandler
	ConnectionManager map[string]Connection
}

func NewService(id string) *Service {
	return &Service{
		LocalNode:       state.CreateLocalNode(id),
		requestHandlers: map[string]RequestHandler{},
	}
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
}
