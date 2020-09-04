package indices

import "github.com/actumn/searchgoose/state"

type ClusterStateService struct {
	IndicesService *Service
}

func NewClusterStateService(indices *Service) *ClusterStateService {
	return &ClusterStateService{
		IndicesService: indices,
	}
}

func (s *ClusterStateService) ApplyClusterState(event state.ClusterChangedEvent) {
	clusterState := event.State

	for _, indexMetadata := range clusterState.Metadata.Indices {
		if _, exists := s.IndicesService.IndexService(indexMetadata.Index.Uuid); !exists {
			indexService := s.IndicesService.CreateIndexService(indexMetadata.Index.Uuid)
			indexService.UpdateMapping(indexMetadata)
			indexService.CreateShard()
		}
	}
}
