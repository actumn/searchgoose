package cluster

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
)

type CreateIndexClusterStateUpdateRequest struct {
	Index    string
	Mappings []byte
	Settings map[string]interface{}
}

type MetadataCreateIndexService struct {
	clusterService    state.ClusterService
	allocationService *AllocationService
}

func NewMetadataCreateIndexService(clusterService state.ClusterService, allocationService *AllocationService) *MetadataCreateIndexService {
	return &MetadataCreateIndexService{
		clusterService:    clusterService,
		allocationService: allocationService,
	}
}

func (s *MetadataCreateIndexService) CreateIndex(req CreateIndexClusterStateUpdateRequest) {
	logrus.Infof("Create index - index name: %s, mapping: %s", req.Index, string(req.Mappings))

	s.clusterService.SubmitStateUpdateTask(func(current state.ClusterState) state.ClusterState {
		return s.applyCreateIndex(current, req)
	})
}

func (s *MetadataCreateIndexService) applyCreateIndex(current state.ClusterState, req CreateIndexClusterStateUpdateRequest) state.ClusterState {
	// prepare Settings
	settings := req.Settings
	var routingNumShards int
	if num, ok := settings["number_of_shards"]; ok {
		routingNumShards = int(num.(float64))
	} else {
		routingNumShards = 3
	}

	// prepare indexMetadata
	indexMetadata := state.IndexMetadata{
		Index: state.Index{
			Name: req.Index,
			Uuid: common.RandomBase64(),
		},
		NumberOfShards: routingNumShards,
		Aliases:        map[string]state.AliasMetadata{},
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
		IndicesLookup: map[string]state.IndexAbstractionAlias{},
	}
	for k, v := range current.Metadata.Indices {
		metadata.Indices[k] = v
	}

	// regenerate routing table using indexMetadata
	shards := map[int]state.IndexShardRoutingTable{}
	for shardNumber := 0; shardNumber < indexMetadata.NumberOfShards; shardNumber++ {
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

	return s.allocationService.reroute(state.ClusterState{
		Name:         current.Name,
		StateUUID:    current.StateUUID,
		Version:      current.Version,
		Nodes:        current.Nodes,
		Metadata:     metadata,
		RoutingTable: routingTable,
	})
}
