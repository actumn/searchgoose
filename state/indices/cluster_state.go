package indices

import (
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
	"sync"
)

type ClusterStateService struct {
	IndicesService *Service
	mux            sync.Mutex
}

func NewClusterStateService(indices *Service) *ClusterStateService {
	return &ClusterStateService{
		IndicesService: indices,
	}
}

func (s *ClusterStateService) ApplyClusterState(event state.ClusterChangedEvent) {
	s.deleteIndices(event)

	s.createIndices(event)
}

func (s *ClusterStateService) deleteIndices(event state.ClusterChangedEvent) {
	//previousState := event.PrevState
	//currentState := event.State
	//localNodeId := currentState.Nodes.LocalNodeId

	for _, idx := range event.IndicesDeleted() {
		if _, existing := s.IndicesService.IndexService(idx.Uuid); existing {
			s.IndicesService.RemoveIndex(idx)
		}
	}
}

func (s *ClusterStateService) createIndices(event state.ClusterChangedEvent) {
	clusterState := event.State

	routingNodes := state.NewRoutingNodes(clusterState)
	localNode := routingNodes.NodesToShards[clusterState.Nodes.LocalNodeId]

	// TODO:: migrate to channel from mutex
	s.mux.Lock()
	for _, shardRouting := range localNode.Shards {
		index := shardRouting.ShardId.Index

		if indexService, exists := s.IndicesService.IndexService(index.Uuid); !exists {
			indexService = s.IndicesService.CreateIndexService(index.Uuid)
			indexMetadata := clusterState.Metadata.Indices[index.Name]
			indexService.UpdateMapping(indexMetadata)
			logrus.Infof("Create new index shard - index name: %s, index uuid: %s, shard number: %d", index.Name, index.Uuid, shardRouting.ShardId.ShardId)
			indexService.CreateShard(shardRouting)
		} else {
			if _, exists := indexService.Shard(shardRouting.ShardId.ShardId); !exists {
				logrus.Infof("Create existing index shard - index name: %s, index uuid: %s, shard number: %d", index.Name, index.Uuid, shardRouting.ShardId.ShardId)
				indexService.CreateShard(shardRouting)
			}
		}
	}
	s.mux.Unlock()
}
