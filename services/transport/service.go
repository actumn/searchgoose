package transport

import "github.com/actumn/searchgoose/services/discovery"

type Service struct {
	LocalNode *discovery.Node
}

type Transport interface {
}
