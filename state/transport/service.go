package transport

import (
	"github.com/actumn/searchgoose/state"
)

type Service struct {
	LocalNode       *state.Node
	requestHandlers map[string]RequestHandler
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

type RequestHandler func(req []byte)

type Transport interface {
}
