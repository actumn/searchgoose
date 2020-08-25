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

	for _, indexMetadata := range clusterState.Metadata.Indices {
		indexService := s.IndicesService.CreateIndexService(indexMetadata.Index.Uuid)
		indexService.UpdateMapping(indexMetadata)
		indexService.CreateShard()
	}
}

type Service struct {
	indices map[string]*index.Service
}

func (s *Service) CreateIndexService(uuid string) *index.Service {
	indexService := index.NewService(uuid)
	s.indices[uuid] = indexService
	return indexService

}
