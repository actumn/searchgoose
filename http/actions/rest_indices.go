package actions

import (
	"github.com/actumn/searchgoose/state/cluster"
	"github.com/actumn/searchgoose/state/indices"
)

type RestIndicesStatsAction struct {
	clusterService *cluster.Service
	indicesService *indices.Service
}

func NewRestIndicesStatsAction(clusterService *cluster.Service, indicesService *indices.Service) *RestIndicesStatsAction {
	return &RestIndicesStatsAction{
		clusterService: clusterService,
		indicesService: indicesService,
	}
}

func (h *RestIndicesStatsAction) Handle(r *RestRequest, reply ResponseListener) {
	// TODO:: reply based on indices service and all cluster shards information
	clusterState := h.clusterService.State()
	indicesMap := map[string]interface{}{}
	for _, index := range clusterState.Metadata.Indices {
		indicesMap[index.Index.Name] = map[string]interface{}{
			"uuid": index.Index.Uuid,
		}
	}
	reply(RestResponse{
		StatusCode: 200,
		Body: map[string]interface{}{
			"_shards": map[string]interface{}{
				"total":      1,
				"successful": 1,
				"failed":     0,
			},
			"_all": map[string]interface{}{
				"primaries": map[string]interface{}{},
				"total":     map[string]interface{}{},
			},
			"indices": indicesMap,
		},
	})
}
