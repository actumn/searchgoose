package indices

import (
	"github.com/actumn/searchgoose/state"
)

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

	routingNodes := state.NewRoutingNodes(clusterState)
	localNode := routingNodes.NodesToShards[clusterState.Nodes.LocalNodeId]

	for _, shardRouting := range localNode.Shards {
		index := shardRouting.ShardId.Index

		if indexService, exists := s.IndicesService.IndexService(index.Uuid); !exists {
			indexService = s.IndicesService.CreateIndexService(index.Uuid)
			indexMetadata := clusterState.Metadata.Indices[index.Name]
			indexService.UpdateMapping(indexMetadata)
			indexService.CreateShard(shardRouting.ShardId.ShardId)
		} else {
			indexService.CreateShard(shardRouting.ShardId.ShardId)
		}
	}
}
