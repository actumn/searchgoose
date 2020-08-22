package indices

import (
	"github.com/actumn/searchgoose/common"
	"github.com/actumn/searchgoose/state"
)

type CreateIndexClusterStateUpdateRequest struct {
	Index    string
	Mappings string
	//Settings Settings
}

type MetadataCreateIndexService struct {
	ClusterService state.ClusterService
}

func (s *MetadataCreateIndexService) CreateIndex(req CreateIndexClusterStateUpdateRequest) {
	s.ClusterService.SubmitStateUpdateTask(func(current state.ClusterState) state.ClusterState {
		return s.applyCreateIndex(current, req)
	})
}

func (s *MetadataCreateIndexService) applyCreateIndex(current state.ClusterState, req CreateIndexClusterStateUpdateRequest) state.ClusterState {
	indexMetadata := state.IndexMetadata{
		State: state.OPEN,
		Index: state.Index{
			Name: req.Index,
			Uuid: common.RandomBase64(),
		},
		RoutingNumShards:   1,
		RoutingNumReplicas: 0,
		Mapping: map[string]state.MappingMetadata{
			"_doc": {
				Type:   "_doc",
				Source: []byte(req.Mappings),
			},
		},
	}

	current.Metadata.Indices[indexMetadata.Index.Name] = indexMetadata

	newState := state.ClusterState{
		Name:      current.Name,
		StateUUID: current.StateUUID,
		Version:   current.Version,
		Nodes:     current.Nodes,
		Metadata:  current.Metadata,
	}
	// TODO:: routing table

	return newState
}

//type AllocationService struct {
//}
