package indices

import (
	"github.com/actumn/searchgoose/env"
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
)

type Service struct {
	Indices map[string]*index.Service
}

func NewService() *Service {
	return &Service{
		Indices: map[string]*index.Service{},
	}
}

func (s *Service) CreateIndexService(uuid string) *index.Service {
	indexService := index.NewService(uuid)
	s.Indices[uuid] = indexService
	return indexService
}

func (s *Service) IndexService(uuid string) (*index.Service, bool) {
	v, ok := s.Indices[uuid]
	return v, ok
}

func (s *Service) RemoveIndex(idx state.Index) {
	//indexName := idx.Name
	_, existing := s.Indices[idx.Uuid]
	if !existing {
		return
	}

	delete(s.Indices, idx.Uuid)
	s.deleteIndexStore(idx)
}

func (s *Service) deleteIndexStore(idx state.Index) {
	// delete index directory here
	if err := env.RemoveContents("./data/" + idx.Uuid); err != nil {
		logrus.Fatal(err)
	}
}
