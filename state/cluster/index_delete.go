package cluster

import (
	"github.com/actumn/searchgoose/state"
	"github.com/sirupsen/logrus"
)

type DeleteIndexClusterStateUpdateRequest struct {
	Indices []state.Index
}

type MetadataDeleteIndexService struct {
	clusterService    state.ClusterService
	allocationService *AllocationService
}

func NewMetadataDeleteIndexService(clusterService state.ClusterService, allocationService *AllocationService) *MetadataDeleteIndexService {
	return &MetadataDeleteIndexService{
		clusterService:    clusterService,
		allocationService: allocationService,
	}
}

func (s *MetadataDeleteIndexService) DeleteIndex(req DeleteIndexClusterStateUpdateRequest) {
	logrus.Infof("Delete index - index name: %s", req.Indices)

	s.clusterService.SubmitStateUpdateTask(func(current state.ClusterState) state.ClusterState {
		indices := map[state.Index]struct{}{}
		for _, index := range req.Indices {
			indices[index] = struct{}{}
		}

		return s.deleteIndices(current, indices)
	})
}

func (s *MetadataDeleteIndexService) deleteIndices(current state.ClusterState, indices map[state.Index]struct{}) state.ClusterState {
	meta := current.Metadata

	routingTable := state.RoutingTable{
		IndicesRouting: map[string]state.IndexRoutingTable{},
	}
	for k, v := range current.RoutingTable.IndicesRouting {
		routingTable.IndicesRouting[k] = v
	}
	metadata := state.Metadata{
		Indices: map[string]state.IndexMetadata{},
	}
	for k, v := range meta.Indices {
		metadata.Indices[k] = v
	}
	for index, _ := range indices {
		indexName := index.Name
		logrus.Info("Deleting index [", index.Name, "/", index.Uuid, "]")
		delete(routingTable.IndicesRouting, indexName)
		delete(metadata.Indices, indexName)
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
