package indices

import (
	"github.com/actumn/searchgoose/index"
	"github.com/actumn/searchgoose/state"
	"log"
)

type ClusterStateService struct {
	IndicesService *Service
}

func (s *ClusterStateService) ApplyClusterState(event state.ClusterChangedEvent) {
	clusterState := event.State

	log.Println(clusterState)

	for name, indexMetadata := range clusterState.Metadata.Indices {
		indexService := s.IndicesService.CreateIndexService()
		indexService.UpdateMapping()
	}
}

type Service struct {
	indices map[string]*index.Service
}

func (s *Service) CreateIndexService() *index.Service {
	//s.indices[uuid] = &index.Service{
	//
	//}
}
