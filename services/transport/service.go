package transport

import (
	"github.com/actumn/searchgoose/services"
)

type Service struct {
	LocalNode *services.Node
}

type Transport interface {
}
