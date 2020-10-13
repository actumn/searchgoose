package indices

import (
	"github.com/actumn/searchgoose/env"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
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

func (s *Service) RemoveIndex(idx state.Index) {
	//indexName := idx.Name
	_, existing := s.indices[idx.Uuid]
	if !existing {
		return
	}

	delete(s.indices, idx.Uuid)
	s.deleteIndexStore(idx)
}

func (s *Service) deleteIndexStore(idx state.Index) {
	// delete index directory here
	if err := env.RemoveContents("./data/" + idx.Uuid); err != nil {
		logrus.Fatal(err)
	}
}
