package cluster

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
)

type CreateIndexClusterStateUpdateRequest struct {
	Index    string
	Mappings []byte
	//Settings Settings
}

type MetadataCreateIndexService struct {
	ClusterService    state.ClusterService
	AllocationService *AllocationService
}

func NewMetadataCreateIndexService(clusterService state.ClusterService, allocationService *AllocationService) *MetadataCreateIndexService {
	return &MetadataCreateIndexService{
		ClusterService:    clusterService,
		AllocationService: allocationService,
	}
}

func (s *MetadataCreateIndexService) CreateIndex(req CreateIndexClusterStateUpdateRequest) {
	logrus.Infof("Create index - index name: %s, mapping: %s", req.Index, string(req.Mappings))

	s.ClusterService.SubmitStateUpdateTask(func(current state.ClusterState) state.ClusterState {
		return s.applyCreateIndex(current, req)
	})
}

func (s *MetadataCreateIndexService) applyCreateIndex(current state.ClusterState, req CreateIndexClusterStateUpdateRequest) state.ClusterState {
	// prepare indexMetadata
	indexMetadata := state.IndexMetadata{
		Index: state.Index{
			Name: req.Index,
			Uuid: common.RandomBase64(),
		},
		RoutingNumShards: 3, // TODO :: get RoutingNumShards from req
		Mapping: map[string]state.MappingMetadata{
			"_doc": {
				Type:   "_doc",
				Source: req.Mappings,
			},
		},
	}

	metadata := state.Metadata{
		Indices: map[string]state.IndexMetadata{
			indexMetadata.Index.Name: indexMetadata,
		},
	}
	for k, v := range current.Metadata.Indices {
		metadata.Indices[k] = v
	}

	// regenerate routing table using indexMetadata
	shards := map[int]state.IndexShardRoutingTable{}
	for shardNumber := 0; shardNumber < indexMetadata.RoutingNumShards; shardNumber++ {
		shards[shardNumber] = state.IndexShardRoutingTable{
			ShardId: state.ShardId{
				Index:   indexMetadata.Index,
				ShardId: shardNumber,
			},
			Primary: state.ShardRouting{
				ShardId: state.ShardId{
					Index:   indexMetadata.Index,
					ShardId: shardNumber,
				},
				CurrentNodeId: "",
				//RelocatingNodeId: "",
				Primary: true,
			},
		}
	}

	routingTable := state.RoutingTable{
		IndicesRouting: map[string]state.IndexRoutingTable{
			indexMetadata.Index.Name: {
				Index:  indexMetadata.Index,
				Shards: shards,
			},
		},
	}
	for k, v := range current.RoutingTable.IndicesRouting {
		routingTable.IndicesRouting[k] = v
	}

	return s.AllocationService.reroute(state.ClusterState{
		Name:         current.Name,
		StateUUID:    current.StateUUID,
		Version:      current.Version,
		Nodes:        current.Nodes,
		Metadata:     metadata,
		RoutingTable: routingTable,
	})
}

//type AllocationService struct {
//}
