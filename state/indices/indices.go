package indices

import (
	"github.com/actumn/searchgoose/index"
)

type Service struct {
	indices map[string]*index.Service
}

func NewService() *Service {
	return &Service{
		indices: map[string]*index.Service{},
	}
}

func (s *Service) CreateIndexService(uuid string) *index.Service {
	indexService := index.NewService(uuid)
	s.indices[uuid] = indexService
	return indexService
}

func (s *Service) IndexService(uuid string) (*index.Service, bool) {
	v, ok := s.indices[uuid]
	return v, ok
}
